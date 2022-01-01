package console

type Settings struct{}

func (s *Settings) SetDefaults() {}

func (s Settings) Validate() (err error) {
	return nil
}
