package substituter

import (
	"encoding/json"
	"fmt"

	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type Settings struct {
	Substitutions []Substitution
}

func (s *Settings) SetDefaults() {
	for i := range s.Substitutions {
		s.Substitutions[i].setDefaults()
	}
}

func (s *Settings) Validate() (err error) {
	for i, substitution := range s.Substitutions {
		err = substitution.validate()
		if err != nil {
			return fmt.Errorf("substitution %d of %d: %w",
				i+1, len(s.Substitutions), err)
		}
	}
	return nil
}

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	if len(s.Substitutions) == 0 {
		return gotree.New("Substitute middleware: disabled")
	}
	node = gotree.New("Substitute middleware settings:")
	substitutionsNode := node.Appendf("Substitutions:")
	for _, substitution := range s.Substitutions {
		substitutionsNode.Append(substitution.String())
	}
	return node
}

func (s *Settings) Read(reader *reader.Reader) (err error) {
	substitutionsStringPtr := reader.Get("MIDDLEWARE_SUBSTITUTER_SUBSTITUTIONS")
	if substitutionsStringPtr == nil {
		return nil
	}

	err = json.Unmarshal([]byte(*substitutionsStringPtr), &s.Substitutions)
	if err != nil {
		return fmt.Errorf("JSON decoding substitutions: %w", err)
	}
	return nil
}
