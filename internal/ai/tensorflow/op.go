package tensorflow

import (
	tf "github.com/wamuir/graft/tensorflow"
)

// AddSoftmax appends a Softmax operation to the graph for the configured model output.
func AddSoftmax(graph *tf.Graph, info *ModelInfo) (*tf.Operation, error) {

	randomName := randomString(10)

	logits := graph.Operation(info.Output.Name).Output(info.Output.OutputIndex)
	reshapeOpSpec := tf.OpSpec{
		Type: "EnsureShape",
		Name: randomString(10),
		Input: []tf.Input{
			logits,
		},
		Attrs: map[string]any{
			"shape": tf.MakeShape(-1, info.Output.NumOutputs),
		},
	}

	// We add this reshape operation becase TF seems unable to infere the input
	// shape for softmax operation, eventhough it is perfectly recoverable by
	// inspecting the models.
	reshapeOp, err := graph.AddOperation(reshapeOpSpec)
	if err != nil {
		return nil, err
	}

	opspec := tf.OpSpec{
		Type: "Softmax",
		Name: randomName,
		Input: []tf.Input{
			reshapeOp.Output(0),
		},
	}

	op, err := graph.AddOperation(opspec)
	if err != nil {
		return nil, err
	}

	info.Output.Name = randomName
	info.Output.OutputIndex = 0

	return op, nil
}
