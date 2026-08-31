package reading

// ingest_stage_test.go is itd-185's ac-2 and ac-10, plus the item-level
// granularity spc-63 records rather than leaving to be discovered at the audit.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/intentdriven/abcd/internal/core/capture"
	"github.com/intentdriven/abcd/internal/fsutil"
)

// withFault arms the staged-write protocol's test seam for one step and restores
// it afterwards. The seam is the only way to execute the crash window from
// inside the process; a protocol whose crash behaviour is never run is a
// protocol nobody has checked.
func withFault(t *testing.T, step string) {
	t.Helper()
	prior := ingestFault
	ingestFault = func(at string) error {
		if at == step {
			return errors.New("injected fault at " + at)
		}
		return nil
	}
	t.Cleanup(func() { ingestFault = prior })
}

// TestRunMetadataLandsLast is ac-2 at both of the protocol's windows.
//
// The criterion is a state, not an instant: after the fault AND the next
// invocation, no reading records and no run metadata are durable for that run,
// and the orphaned stage has been named and cleared. Injecting at the earlier
// window alone would leave the later one — where the ledger records have already
// landed — untested, which is the window that actually needs a rollback.
func TestRunMetadataLandsLast(t *testing.T) {
	for _, step := range []string{faultAfterStage, faultAfterLedger} {
		t.Run(step, func(t *testing.T) {
			f := newIngestFixture(t, "detection")
			doc := f.payload(2)

			withFault(t, step)
			if _, err := f.ingest(doc); err == nil {
				t.Fatal("the injected fault did not stop the ingest")
			}
			ingestFault = nil

			// The commit marker is the whole claim: whatever else survived the
			// fault, the run never happened.
			if f.exists(ReadingsRecordDir + "/" + f.runID + "/" + RunFileName) {
				t.Fatal("the run metadata landed despite a fault before it")
			}
			if !f.exists(IngestStageDir + "/" + f.runID) {
				t.Fatal("the fault left no stage, so a later invocation has no evidence to find")
			}

			// The next invocation names the orphan and clears it, and the
			// rollback leaves nothing durable for the run.
			f.parkRun("rdg-2608310000000009", "detection", AssemblerVersion)
			next := f.payload(1)
			next["run_id"] = "rdg-2608310000000009"
			next["manifest_sha256"] = f.manifestHashOf("rdg-2608310000000009")
			res := f.mustIngest(next)

			named := false
			for _, run := range res.ClearedStages {
				if run == f.runID {
					named = true
				}
			}
			if !named {
				t.Errorf("the next invocation cleared %v, which does not name the orphaned run %s",
					res.ClearedStages, f.runID)
			}
			f.nothingDurable(f.runID)
		})
	}
}

// TestOrphanedStageIsReportedAndCleared is the sweep on its own, over an orphan
// this test constructs rather than one an injected fault produced — so the sweep
// is proved against the state, not only against the path that reaches it.
//
// Three things must be true at once: the orphan is named, its half-written
// records are gone, and a directory the sweep has no business in is untouched.
func TestOrphanedStageIsReportedAndCleared(t *testing.T) {
	f := newIngestFixture(t, "detection")
	orphan := "rdg-2608310000000003"

	f.write(IngestStageDir+"/"+orphan+"/"+stageFileName,
		[]byte(`{"_type":"`+StageType+`","run_id":"`+orphan+`","records":["rdi-2608310000000004"]}`))
	f.write(".abcd/work/issues/readings/"+orphan+"/rdi-2608310000000004.md", []byte("---\nid: rdi-2608310000000004\n---\n"))
	// A file the sweep must not touch: it is not a reading record, and a
	// rollback that cleared the directory would take it with the rest.
	f.write(".abcd/work/issues/readings/"+orphan+"/NOTES.md", []byte("a person put this here\n"))
	// A stage directory that is not a run id at all. The sweep leaves it alone
	// rather than deleting somebody's file on a guess.
	f.write(IngestStageDir+"/notes.txt", []byte("not a run\n"))

	res := f.mustIngest(f.payload(1))
	if len(res.ClearedStages) != 1 || res.ClearedStages[0] != orphan {
		t.Fatalf("the sweep cleared %v, want exactly [%s]", res.ClearedStages, orphan)
	}
	if f.exists(IngestStageDir + "/" + orphan) {
		t.Error("the orphaned stage survived the sweep")
	}
	if f.exists(".abcd/work/issues/readings/" + orphan + "/rdi-2608310000000004.md") {
		t.Error("the orphaned run's reading record survived the rollback")
	}
	// A delete in the COMMITTED tier is reported by id. "Cleared an orphaned
	// stage" does not tell an operator that records left the ledger with it.
	if len(res.RolledBack) != 1 || res.RolledBack[0] != "rdi-2608310000000004" {
		t.Errorf("the sweep rolled back %v; it removed rdi-2608310000000004 from the ledger and has to "+
			"say so", res.RolledBack)
	}
	if !f.exists(".abcd/work/issues/readings/" + orphan + "/NOTES.md") {
		t.Error("the rollback removed a file that is not a reading record")
	}
	if !f.exists(IngestStageDir + "/notes.txt") {
		t.Error("the sweep removed a stage entry that is not a run id")
	}
}

// TestOrphanSweepLeavesACommittedRunAlone is the sweep's other half, and it is
// the one that makes the rollback safe to have: a stage beside a run whose
// commit marker DID land is a leftover from a crash after the marker. The run
// happened, and only the stage goes.
func TestOrphanSweepLeavesACommittedRunAlone(t *testing.T) {
	f := newIngestFixture(t, "detection")
	res := f.mustIngest(f.payload(2))
	if len(res.Records) != 2 {
		t.Fatalf("the first run landed %d of 2 records", len(res.Records))
	}
	before := f.ledgerRecords(f.runID)

	// A stage reappears beside the completed run.
	f.write(IngestStageDir+"/"+f.runID+"/"+stageFileName,
		[]byte(`{"_type":"`+StageType+`","run_id":"`+f.runID+`","records":[]}`))

	f.parkRun("rdg-2608310000000005", "detection", AssemblerVersion)
	next := f.payload(1)
	next["run_id"] = "rdg-2608310000000005"
	next["manifest_sha256"] = f.manifestHashOf("rdg-2608310000000005")
	f.mustIngest(next)

	if f.exists(IngestStageDir + "/" + f.runID) {
		t.Error("the leftover stage survived")
	}
	if got := f.ledgerRecords(f.runID); len(got) != len(before) {
		t.Errorf("the sweep rolled back a COMMITTED run: %v became %v", before, got)
	}
	if !f.exists(ReadingsRecordDir + "/" + f.runID + "/" + RunFileName) {
		t.Error("the sweep removed a committed run's commit marker")
	}
}

// TestItemLevelViolationLandsTheRest is the refusal granularity spc-63 records:
// an item-level violation refuses that item and lands the rest, naming the
// refused item's ordinal and the rule.
func TestItemLevelViolationLandsTheRest(t *testing.T) {
	f := newIngestFixture(t, "detection")
	doc := f.payload(3)
	doc["items"].([]any)[1].(map[string]any)[PatternField] = ""

	res := f.mustIngest(doc)
	if len(res.Records) != 2 {
		t.Fatalf("an item-level refusal landed %d records, want the 2 survivors", len(res.Records))
	}
	if len(res.RefusedItems) != 1 {
		t.Fatalf("reported %d refusals, want 1: %v", len(res.RefusedItems), res.RefusedItems)
	}
	if res.RefusedItems[0].Ordinal != 2 {
		t.Errorf("the refusal names ordinal %d, want 2", res.RefusedItems[0].Ordinal)
	}
	if strings.TrimSpace(res.RefusedItems[0].Rule) == "" {
		t.Error("the refusal names no rule")
	}

	// And it is durable on the run record, so the refusal outlives the render.
	run := f.readRunRecord(f.runID)
	if len(run.RefusedItems) != 1 || run.RefusedItems[0].Ordinal != 2 {
		t.Errorf("the run record carries refusals %v", run.RefusedItems)
	}
	if len(run.Records) != 2 {
		t.Errorf("the run record names %d records, want 2", len(run.Records))
	}
}

// TestListLevelRefusalWritesRefusalRecordOnly is ac-10: a run refused at list
// level leaves a refusal record carrying the run metadata and the named reason,
// and no reading records.
func TestListLevelRefusalWritesRefusalRecordOnly(t *testing.T) {
	f := newIngestFixture(t, "detection")
	doc := f.payload(3)
	doc["regime"] = RegimeEvaluative

	res, err := f.ingest(doc)
	if err == nil {
		t.Fatal("a list-level violation was accepted")
	}
	if res.RefusalPath == "" {
		t.Fatal("the refusal reported no record path")
	}

	rec := f.readRefusalRecord(f.runID)
	if rec.Type != RefusalType {
		t.Errorf("the refusal record states _type %q", rec.Type)
	}
	if rec.RunID != f.runID || rec.Position != f.position || rec.Regime != f.regime {
		t.Errorf("the refusal record carries the wrong run metadata: %+v", rec)
	}
	if rec.ManifestSHA256 != f.manifestHash || rec.TargetCommit == "" {
		t.Errorf("the refusal record does not carry the run's manifest reference: %+v", rec)
	}
	// The WHOLE reason, not a prefix of it that happens to contain the word:
	// the recorded reason is what the operator was told, and a record that
	// carries the first hundred runes of it carries a different claim.
	if !strings.Contains(rec.Reason, RegimeEvaluative) || !strings.Contains(rec.Reason, "refuses the run") {
		t.Errorf("the refusal record does not carry the whole named reason: %q", rec.Reason)
	}

	if got := f.ledgerRecords(f.runID); len(got) != 0 {
		t.Errorf("the refused run wrote %v into the reading-record family", got)
	}
	if f.exists(ReadingsRecordDir + "/" + f.runID + "/" + RunFileName) {
		t.Error("a refused run wrote a commit marker")
	}
	if f.exists(IngestStageDir + "/" + f.runID) {
		t.Error("a refused run left a stage behind")
	}

	// A rerun is a NEW run with a new run id, never an amendment, and the
	// refusal record of the refused one stands.
	f.parkRun("rdg-2608310000000006", "detection", AssemblerVersion)
	next := f.payload(1)
	next["run_id"] = "rdg-2608310000000006"
	next["manifest_sha256"] = f.manifestHashOf("rdg-2608310000000006")
	f.mustIngest(next)
	if !f.exists(ReadingsRecordDir + "/" + f.runID + "/" + RefusalFileName) {
		t.Error("a later run erased the earlier run's refusal record")
	}
}

// TestRunIDNeverBuildsAPathBeforeItIsChecked is the trust boundary stated as a
// property: a run id is the one payload value a path is built from, so a
// traversal id must be refused before any file is opened — and must therefore
// leave nothing outside the repository behind.
func TestRunIDNeverBuildsAPathBeforeItIsChecked(t *testing.T) {
	f := newIngestFixture(t, "detection")
	outside := filepath.Join(f.root, "..", "escaped")
	for _, id := range []string{
		"rdg-../../../escaped", "../escaped", "rdg-2608310000000001/../../escaped",
		"rdg-" + strings.Repeat("9", 400),
	} {
		doc := f.payload(1)
		doc["run_id"] = id
		if _, err := f.ingest(doc); err == nil {
			t.Errorf("run_id %q was accepted", id)
		}
	}
	if _, err := os.Stat(outside); !os.IsNotExist(err) {
		t.Errorf("a traversal run id produced %s", outside)
	}
}

// TestARerunOfACommittedRunIsRefused holds "a rerun is a NEW run with a new run
// id, never an amendment" as a check rather than a sentence.
//
// The run id is payload-chosen. Without the check a second ingest lands a second
// batch of records beside the first and rewrites run.json to name only the
// second — and the first batch is then unreachable from any run record AND
// beyond every later sweep, because the rollback bails whenever a commit marker
// exists. The refusal record is guarded for the same reason in the other
// direction: a rerun must not overwrite the refusal of the run it repeats.
func TestARerunOfACommittedRunIsRefused(t *testing.T) {
	t.Run("after a commit marker", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		doc := f.payload(2)
		first := f.mustIngest(doc)
		if len(first.Records) != 2 {
			t.Fatalf("the first run landed %d of 2", len(first.Records))
		}

		if _, err := f.ingest(doc); err == nil {
			t.Fatal("a run that already committed was ingested again")
		} else if !strings.Contains(err.Error(), "never an amendment") {
			t.Errorf("the refusal does not state the rule: %v", err)
		}
		if got := f.ledgerRecords(f.runID); len(got) != 2 {
			t.Errorf("the ledger holds %v; the rerun duplicated the run's records", got)
		}
		run := f.readRunRecord(f.runID)
		if len(run.Records) != 2 {
			t.Errorf("the run record names %d records", len(run.Records))
		}
	})

	t.Run("after a refusal record", func(t *testing.T) {
		f := newIngestFixture(t, "detection")
		bad := f.payload(1)
		bad["regime"] = RegimeEvaluative
		if _, err := f.ingest(bad); err == nil {
			t.Fatal("a regime mismatch was accepted")
		}
		before := f.readRefusalRecord(f.runID)

		if _, err := f.ingest(f.payload(1)); err == nil {
			t.Fatal("a run that was already refused was ingested again")
		}
		if after := f.readRefusalRecord(f.runID); after.Reason != before.Reason {
			t.Error("the rerun overwrote the refusal record of the run it repeated")
		}
		if f.exists(ReadingsRecordDir + "/" + f.runID + "/" + RunFileName) {
			t.Error("a rerun of a refused run wrote a commit marker beside its refusal")
		}
	})

	// And a DIFFERENT run in the same repository still lands: the rule is about
	// one run id, not about the repository having ingested before.
	f := newIngestFixture(t, "detection")
	f.mustIngest(f.payload(1))
	if res := f.mustIngest(f.nextRun(f.payload(1))); len(res.Records) != 1 {
		t.Fatalf("a second, distinct run landed %d record(s)", len(res.Records))
	}
}

// TestTheStageLockIsHeldAcrossTheSweepAndTheWrite is the peer-invocation guard.
//
// The sweep deletes committed reading records, and its only test for an orphan
// is a stage with no commit marker beside it — which is exactly what a LIVE
// ingest looks like between its ledger write and its marker. A second
// invocation would therefore roll the first one back mid-flight, and the first
// would then write a run record naming records that no longer exist and exit 0.
//
// The lock is probed from inside the fault seam, at the very window the race
// needs, with a zero timeout: a race driven by two real processes would be
// timing-dependent, and this asserts the property the race rests on instead.
func TestTheStageLockIsHeldAcrossTheSweepAndTheWrite(t *testing.T) {
	f := newIngestFixture(t, "detection")
	lock := filepath.Join(f.root, filepath.FromSlash(IngestStageDir), stageLockName)

	probed := map[string]bool{}
	prior := ingestFault
	ingestFault = func(at string) error {
		probed[at] = true
		err := fsutil.WithFileLock(lock, 0, func() error { return nil })
		if !errors.Is(err, fsutil.ErrLockContention) {
			t.Errorf("at %s the stage lock was free (%v); a peer invocation could sweep this run's "+
				"records out from under it", at, err)
		}
		return nil
	}
	t.Cleanup(func() { ingestFault = prior })

	f.mustIngest(f.payload(2))
	for _, at := range []string{faultAfterStage, faultAfterLedger} {
		if !probed[at] {
			t.Errorf("the %s window was never reached, so the lock was never probed there", at)
		}
	}

	// And the lock is RELEASED: the next ingest is not blocked by the last.
	f.mustIngest(f.nextRun(f.payload(1)))
	if err := fsutil.WithFileLock(lock, 0, func() error { return nil }); err != nil {
		t.Errorf("the stage lock was not released after the ingest returned: %v", err)
	}
}

// TestTheLedgerLockIsHeldWhileTheSweepUnlinks.
//
// The stage lock serialises ingest against ingest and says nothing about
// core/capture, whose own verbs READ these records. The sweep unlinks committed
// reading records, so without capture's ledger lock a concurrent disposition or
// promote can be reading a record as it disappears.
//
// The lock is probed from inside the unlink, for the reason the stage-lock case
// gives: a lock held across a window cannot be observed from outside the
// process, and a race driven by two real processes would be timing-dependent.
// The probe goes through capture's own exported helper rather than rebuilding
// the lock path, so it is the same lock by construction.
func TestTheLedgerLockIsHeldWhileTheSweepUnlinks(t *testing.T) {
	f := newIngestFixture(t, "detection")
	orphan := "rdg-2608310000000021"
	f.write(IngestStageDir+"/"+orphan+"/"+stageFileName,
		[]byte(`{"_type":"`+StageType+`","run_id":"`+orphan+`","records":[]}`))
	f.write(".abcd/work/issues/readings/"+orphan+"/rdi-2608310000000022.md",
		[]byte("---\nid: rdi-2608310000000022\n---\n"))

	probed := false
	prior := ingestFault
	ingestFault = func(at string) error {
		if at != faultDuringRollback {
			return nil
		}
		probed = true
		// capture's own timeout bounds this; it returns contention rather than
		// hanging, which is what makes the assertion deterministic.
		if err := capture.WithLedgerLock(f.root, func() error { return nil }); err == nil {
			t.Error("the ledger lock was free while the sweep unlinked committed records; a concurrent " +
				"disposition or promote could be reading one as it disappears")
		}
		return nil
	}
	t.Cleanup(func() { ingestFault = prior })

	f.mustIngest(f.payload(1))
	if !probed {
		t.Fatal("the unlink was never reached, so the lock was never probed")
	}
	if f.exists(".abcd/work/issues/readings/" + orphan + "/rdi-2608310000000022.md") {
		t.Error("the orphaned run's record survived the rollback")
	}
	// And it is released: the next ledger mutation is not blocked by the sweep.
	if err := capture.WithLedgerLock(f.root, func() error { return nil }); err != nil {
		t.Errorf("the ledger lock was not released after the sweep: %v", err)
	}
}
