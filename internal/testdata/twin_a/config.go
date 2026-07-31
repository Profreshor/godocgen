package twina

package twin_a

type Config struct{}

// BelongsToA must stay in twin_a
func (c *Config) BelongsToA() {}
