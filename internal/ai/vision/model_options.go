package vision

// ModelOptions represents additional model parameters listed in the documentation.
// Comments note which engines currently honor each field.
type ModelOptions struct {
	Temperature      float64  `yaml:"Temperature,omitempty" json:"temperature,omitempty"`            // Ollama, OpenAI
	TopK             int      `yaml:"TopK,omitempty" json:"top_k,omitempty"`                         // Ollama
	TopP             float64  `yaml:"TopP,omitempty" json:"top_p,omitempty"`                         // Ollama, OpenAI
	MinP             float64  `yaml:"MinP,omitempty" json:"min_p,omitempty"`                         // Ollama
	TypicalP         float64  `yaml:"TypicalP,omitempty" json:"typical_p,omitempty"`                 // Ollama
	TfsZ             float64  `yaml:"TfsZ,omitempty" json:"tfs_z,omitempty"`                         // Ollama
	Seed             int      `yaml:"Seed,omitempty" json:"seed,omitempty"`                          // Ollama
	NumKeep          int      `yaml:"NumKeep,omitempty" json:"num_keep,omitempty"`                   // Ollama
	RepeatLastN      int      `yaml:"RepeatLastN,omitempty" json:"repeat_last_n,omitempty"`          // Ollama
	RepeatPenalty    float64  `yaml:"RepeatPenalty,omitempty" json:"repeat_penalty,omitempty"`       // Ollama
	PresencePenalty  float64  `yaml:"PresencePenalty,omitempty" json:"presence_penalty,omitempty"`   // OpenAI
	FrequencyPenalty float64  `yaml:"FrequencyPenalty,omitempty" json:"frequency_penalty,omitempty"` // OpenAI
	PenalizeNewline  bool     `yaml:"PenalizeNewline,omitempty" json:"penalize_newline,omitempty"`   // Ollama
	Stop             []string `yaml:"Stop,omitempty" json:"stop,omitempty"`                          // Ollama, OpenAI
	Mirostat         int      `yaml:"Mirostat,omitempty" json:"mirostat,omitempty"`                  // Ollama
	MirostatTau      float64  `yaml:"MirostatTau,omitempty" json:"mirostat_tau,omitempty"`           // Ollama
	MirostatEta      float64  `yaml:"MirostatEta,omitempty" json:"mirostat_eta,omitempty"`           // Ollama
	NumPredict       int      `yaml:"NumPredict,omitempty" json:"num_predict,omitempty"`             // Ollama
	MaxOutputTokens  int      `yaml:"MaxOutputTokens,omitempty" json:"max_output_tokens,omitempty"`  // Ollama, OpenAI
	ForceJson        bool     `yaml:"ForceJson,omitempty" json:"force_json,omitempty"`               // Ollama, OpenAI
	SchemaVersion    string   `yaml:"SchemaVersion,omitempty" json:"schema_version,omitempty"`       // Ollama, OpenAI
	CombineOutputs   string   `yaml:"CombineOutputs,omitempty" json:"combine_outputs,omitempty"`     // OpenAI
	Detail           string   `yaml:"Detail,omitempty" json:"detail,omitempty"`                      // OpenAI
	NumCtx           int      `yaml:"NumCtx,omitempty" json:"num_ctx,omitempty"`                     // Ollama, OpenAI
	NumThread        int      `yaml:"NumThread,omitempty" json:"num_thread,omitempty"`               // Ollama
	NumBatch         int      `yaml:"NumBatch,omitempty" json:"num_batch,omitempty"`                 // Ollama
	NumGpu           int      `yaml:"NumGpu,omitempty" json:"num_gpu,omitempty"`                     // Ollama
	MainGpu          int      `yaml:"MainGpu,omitempty" json:"main_gpu,omitempty"`                   // Ollama
	LowVram          bool     `yaml:"LowVram,omitempty" json:"low_vram,omitempty"`                   // Ollama
	VocabOnly        bool     `yaml:"VocabOnly,omitempty" json:"vocab_only,omitempty"`               // Ollama
	UseMmap          bool     `yaml:"UseMmap,omitempty" json:"use_mmap,omitempty"`                   // Ollama
	UseMlock         bool     `yaml:"UseMlock,omitempty" json:"use_mlock,omitempty"`                 // Ollama
	Numa             bool     `yaml:"Numa,omitempty" json:"numa,omitempty"`                          // Ollama
}
