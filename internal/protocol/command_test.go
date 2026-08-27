package protocol

import "testing"

func TestCommandBatchValidationAndClone(t *testing.T) {
	command, err := NewCommand(1, "test.intent", map[string]any{"value": 7})
	if err != nil {
		t.Fatal(err)
	}
	batch := CommandBatch{
		SchemaVersion: CommandSchemaVersion,
		GameID:        "game-1",
		SeatID:        1,
		Turn:          3,
		BaseRevision:  9,
		Commands:      []Command{command},
	}
	if err := batch.Validate(); err != nil {
		t.Fatal(err)
	}

	clone := CloneCommandBatch(batch)
	clone.Commands[0].Payload[0] = '['
	if string(clone.Commands[0].Payload) == string(batch.Commands[0].Payload) {
		t.Fatal("clone shares command payload backing storage")
	}
}

func TestCommandBatchRejectsUnstableSequence(t *testing.T) {

	batch := CommandBatch{
		SchemaVersion: CommandSchemaVersion,
		GameID:        "game-1",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands: []Command{{
			SchemaVersion: CommandSchemaVersion,
			Sequence:      2,
			Kind:          "test.intent",
		}},
	}
	if err := batch.Validate(); err == nil {
		t.Fatal("expected non-contiguous sequence to fail")
	}
}
