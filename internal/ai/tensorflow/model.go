package tensorflow

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/pkg/clean"
)

// SavedModel loads a saved TensorFlow model from the specified path.
func SavedModel(modelPath string, tags []string) (model *tf.SavedModel, err error) {
	log.Infof("tensorflow: loading %s", clean.Log(filepath.Base(modelPath)))

	if len(tags) == 0 {
		tags = []string{"serve"}
	}

	return tf.LoadSavedModel(modelPath, tags, nil)
}

// GuessInputAndOutput tries to inspect a loaded saved model to build the
// ModelInfo struct
func GuessInputAndOutput(model *tf.SavedModel) (input *PhotoInput, output *ModelOutput, err error) {
	if model == nil {
		return nil, nil, fmt.Errorf("tensorflow: GuessInputAndOutput received a nil input")
	}

	modelOps := model.Graph.Operations()

	for i := range modelOps {
		if strings.HasPrefix(modelOps[i].Type(), "Placeholder") &&
			modelOps[i].NumOutputs() == 1 &&
			modelOps[i].Output(0).Shape().NumDimensions() == 4 {

			shape := modelOps[i].Output(0).Shape()

			var comps []ShapeComponent
			if shape.Size(3) == ExpectedChannels {
				comps = []ShapeComponent{ShapeBatch, ShapeHeight, ShapeWidth, ShapeColor}
			} else if shape.Size(1) == ExpectedChannels { // check the channels are 3
				comps = []ShapeComponent{ShapeBatch, ShapeColor, ShapeHeight, ShapeWidth, ShapeColor}
			}

			if comps != nil {
				input = &PhotoInput{
					Name:   modelOps[i].Name(),
					Height: shape.Size(1),
					Width:  shape.Size(2),
					Shape:  comps,
				}
			}
		} else if (modelOps[i].Type() == "Softmax" || strings.HasPrefix(modelOps[i].Type(), "StatefulPartitionedCall")) &&
			modelOps[i].NumOutputs() == 1 && modelOps[i].Output(0).Shape().NumDimensions() == 2 {
			output = &ModelOutput{
				Name:       modelOps[i].Name(),
				NumOutputs: modelOps[i].Output(0).Shape().Size(1),
			}
		}

		if input != nil && output != nil {
			return
		}
	}

	return nil, nil, fmt.Errorf("could not guess the inputs and outputs")
}

func GetInputAndOutputFromSavedModel(model *tf.SavedModel) (*PhotoInput, *ModelOutput, error) {
	if model == nil {
		return nil, nil, fmt.Errorf("GetInputAndOutputFromSavedModel: nil input")
	}

	log.Debugf("tensorflow: found %d signatures", len(model.Signatures))
	for k, v := range model.Signatures {
		var photoInput *PhotoInput
		var modelOutput *ModelOutput

		inputs := v.Inputs
		outputs := v.Outputs

		if len(inputs) >= 1 && len(outputs) >= 1 {
			for _, inputTensor := range inputs {
				if inputTensor.Shape.NumDimensions() == 4 {
					var comps []ShapeComponent
					if inputTensor.Shape.Size(3) == ExpectedChannels {
						comps = []ShapeComponent{ShapeBatch, ShapeHeight, ShapeWidth, ShapeColor}
					} else if inputTensor.Shape.Size(1) == ExpectedChannels { // check the channels are 3
						comps = []ShapeComponent{ShapeBatch, ShapeColor, ShapeHeight, ShapeWidth}
					} else {
						log.Debugf("tensorflow: shape %d", inputTensor.Shape.Size(1))
					}

					if comps == nil {
						log.Warnf("tensorflow: skipping signature %v because we could not find the color component", k)
					} else {
						var inputIdx = 0
						var err error

						inputName, inputIndex, found := strings.Cut(inputTensor.Name, ":")
						if found {
							inputIdx, err = strconv.Atoi(inputIndex)
							if err != nil {
								return nil, nil, fmt.Errorf("could not parse index %s (%s)", inputIndex, clean.Error(err))
							}
						}

						photoInput = &PhotoInput{
							Name:        inputName,
							OutputIndex: inputIdx,
							Height:      inputTensor.Shape.Size(1),
							Width:       inputTensor.Shape.Size(2),
							Shape:       comps,
						}
					}

					break
				}
			}

			for outputVarName, outputTensor := range outputs {
				var err error
				var outputIdx int
				if outputTensor.Shape.NumDimensions() == 2 {
					outputName, outputIndex, found := strings.Cut(outputTensor.Name, ":")
					if found {
						outputIdx, err = strconv.Atoi(outputIndex)
						if err != nil {
							return nil, nil, fmt.Errorf("could not parse index %s (%s)", outputIndex, clean.Error(err))
						}
					}

					modelOutput = &ModelOutput{
						Name:          outputName,
						OutputIndex:   outputIdx,
						NumOutputs:    outputTensor.Shape.Size(1),
						OutputsLogits: strings.Contains(outputVarName, "logits"),
					}
					break
				}
			}
		}

		if photoInput != nil && modelOutput != nil {
			return photoInput, modelOutput, nil
		}
	}
	return nil, nil, fmt.Errorf("GetInputAndOutputFromSignature: could not find valid signatures")
}
