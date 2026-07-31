package twin_b

type Config struct{}

// BelongsToB must stay in twin_b
func (c *Config) BelongsToB() {}
