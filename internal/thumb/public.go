package thumb

// Public represents public thumbnail URLs with dimensions.
type Public struct {
	Fit720  Thumb `json:"fit_720"`
	Fit1280 Thumb `json:"fit_1280"`
	Fit1920 Thumb `json:"fit_1920"`
	Fit2560 Thumb `json:"fit_2560"`
	Fit4096 Thumb `json:"fit_4096"`
	Fit7680 Thumb `json:"fit_7680"`
}
