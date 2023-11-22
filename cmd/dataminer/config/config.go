package config

type Config struct {
	// miner settings.
	// How far back a window (in 100ms frames) should the miner save each file
	OrderbookFrames int `yaml:"OrderbookFrames"`
	// How many frames should the miner should reserve to successfully change the current history
	ChangeoverFrames int `yaml:"ChangeoverFrames"`

	// buffer size for packager.
	Buffer int `yaml:"Buffer"`

	// formatring either bin or json
	Format string `yaml:"Format"`

	// Symbols to mine
	Symbols []string `yaml:"Symbols"`

	// local location to save. If given then the dataminer will save locally to this location
	Filepath string `yaml:"Filepath"`

	// AWS settings
	// 1 = true, 0 = false
	Aws int `yaml:"Aws"`

	Key    string `yaml:"Key"`
	Secret string `yaml:"Secret"`
	Region string `yaml:"Region"`

	BucketName string `yaml:"BucketName"`

	// logger settings
	// 1 = true, 0 = false
	Logger int `yaml:"Logger"`
	//save logs where
	LogFilepath string `yaml:"LogFilepath"`
}
