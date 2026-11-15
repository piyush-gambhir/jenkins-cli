package output

import (
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

type YAMLFormatter struct {
	Writer io.Writer
}

func (f *YAMLFormatter) Format(data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	_, err = fmt.Fprint(f.Writer, string(b))
	return err
}
