package schema

import (
	"errors"
	"fmt"
)

func Validate(def *MockDefinition) error {
	switch def.Protocol {
	case "http":
		if len(def.Routes) == 0 {
			return errors.New("⚠️ 'routes' must be defined for HTTP protocol")
		}
	case "tcp", "ws":
		if def.OnMessage == nil {
			return errors.New("⚠️ 'onMessage' must be defined for TCP/WS protocol")
		}
	case "sftp":
		if len(def.Files) == 0 {
			return errors.New("⚠️ 'files' must be defined for SFTP protocol")
		}
		if def.SFTPAuth == nil {
			return errors.New("⚠️ 'sftpAuth' must be defined for SFTP protocol")
		}
		if def.SFTPAuth.Username == "" || def.SFTPAuth.Password == "" {
			return errors.New("⚠️ 'sftpAuth' must include username and password")
		}
	default:
		return fmt.Errorf("❌ unsupported protocol: %s", def.Protocol)
	}
	return nil
}
