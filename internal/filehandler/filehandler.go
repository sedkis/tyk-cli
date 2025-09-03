package filehandler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SupportedExtensions lists the file extensions we support
var SupportedExtensions = []string{".yaml", ".yml", ".json"}

// FileType represents the type of file content
type FileType int

const (
	FileTypeJSON FileType = iota
	FileTypeYAML
)

// FileInfo contains information about a loaded file
type FileInfo struct {
	Path     string
	Type     FileType
	Content  map[string]interface{}
	RawBytes []byte
}

// LoadFile loads and parses a file, automatically detecting its format
func LoadFile(filePath string) (*FileInfo, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Determine file type from extension
	fileType := getFileType(filePath)
	
	// Parse content based on file type
	var parsedContent map[string]interface{}
	switch fileType {
	case FileTypeJSON:
		if err := json.Unmarshal(content, &parsedContent); err != nil {
			return nil, fmt.Errorf("failed to parse JSON file %s: %w", filePath, err)
		}
	case FileTypeYAML:
		if err := yaml.Unmarshal(content, &parsedContent); err != nil {
			return nil, fmt.Errorf("failed to parse YAML file %s: %w", filePath, err)
		}
	default:
		return nil, fmt.Errorf("unsupported file type for %s (supported: %v)", filePath, SupportedExtensions)
	}

	return &FileInfo{
		Path:     filePath,
		Type:     fileType,
		Content:  parsedContent,
		RawBytes: content,
	}, nil
}

// LoadFileAsRawJSON loads a file and converts it to raw JSON bytes
func LoadFileAsRawJSON(filePath string) (json.RawMessage, error) {
	fileInfo, err := LoadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Convert parsed content back to JSON
	jsonBytes, err := json.Marshal(fileInfo.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	return json.RawMessage(jsonBytes), nil
}

// SaveFile saves content to a file in the specified format
func SaveFile(filePath string, content map[string]interface{}) error {
	fileType := getFileType(filePath)
	
	var data []byte
	var err error

	switch fileType {
	case FileTypeJSON:
		data, err = json.MarshalIndent(content, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	case FileTypeYAML:
		data, err = yaml.Marshal(content)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file type for %s (supported: %v)", filePath, SupportedExtensions)
	}

	// Create directory if it doesn't exist
	if dir := filepath.Dir(filePath); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// ValidateFilePath checks if a file path is valid and supported
func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	for _, supportedExt := range SupportedExtensions {
		if ext == supportedExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported file extension %s (supported: %v)", ext, SupportedExtensions)
}

// GetOASVersion extracts the OpenAPI version from parsed content
func GetOASVersion(content map[string]interface{}) string {
	if openapi, ok := content["openapi"].(string); ok {
		return openapi
	}
	if swagger, ok := content["swagger"].(string); ok {
		return swagger
	}
	return ""
}

// GetOASInfo extracts info section from OAS content
func GetOASInfo(content map[string]interface{}) map[string]interface{} {
	if info, ok := content["info"].(map[string]interface{}); ok {
		return info
	}
	return nil
}

// GetOASInfoVersion extracts the version from OAS info section
func GetOASInfoVersion(content map[string]interface{}) string {
	if info := GetOASInfo(content); info != nil {
		if version, ok := info["version"].(string); ok {
			return version
		}
	}
	return ""
}

// GetOASTitle extracts the title from OAS info section
func GetOASTitle(content map[string]interface{}) string {
	if info := GetOASInfo(content); info != nil {
		if title, ok := info["title"].(string); ok {
			return title
		}
	}
	return ""
}

// getFileType determines file type from extension
func getFileType(filePath string) FileType {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return FileTypeJSON
	case ".yaml", ".yml":
		return FileTypeYAML
	default:
		return FileTypeJSON // Default fallback
	}
}

// ConvertToJSON converts any supported file content to JSON
func ConvertToJSON(content map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(content, "", "  ")
}

// ConvertToYAML converts any supported file content to YAML
func ConvertToYAML(content map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(content)
}