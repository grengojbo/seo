package seo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qor/admin"
	"github.com/qor/media/media_library"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

// QorSEOSettingInterface support customize Seo model
type QorSEOSettingInterface interface {
	GetName() string
	SetName(string)
	GetSEOSetting() Setting
	GetGlobalSetting() map[string]string
	SetGlobalSetting(map[string]string)
	GetSEOType() string
	SetSEOType(string)
	GetIsGlobalSEO() bool
	SetIsGlobalSEO(bool)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
	SetCollection(*Collection)
}

// QorSEOSetting default seo model
type QorSEOSetting struct {
	Name        string `gorm:"primary_key"`
	Setting     Setting
	IsGlobalSEO bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	collection *Collection
}

// Setting defined meta's attributes
type Setting struct {
	Title             string `gorm:"size:4294967295"`
	Description       string
	Keywords          string
	Type              string
	OpenGraphURL      string
	OpenGraphType     string
	OpenGraphImage    media_library.MediaBox
	OpenGraphMetadata []OpenGraphMetadata
	EnabledCustomize  bool
	GlobalSetting     map[string]string
}

// OpenGraphMetadata open graph meta data
type OpenGraphMetadata struct {
	Property string
	Content  string
}

// GetSEOSetting get seo setting
func (s QorSEOSetting) GetSEOSetting() Setting {
	return s.Setting
}

// GetName get QorSeoSetting's name
func (s QorSEOSetting) GetName() string {
	return s.Name
}

// SetName set QorSeoSetting's name
func (s *QorSEOSetting) SetName(name string) {
	s.Name = name
}

// GetSEOType get QorSeoSetting's type
func (s QorSEOSetting) GetSEOType() string {
	return s.Setting.Type
}

// SetSEOType set QorSeoSetting's type
func (s *QorSEOSetting) SetSEOType(t string) {
	s.Setting.Type = t
}

// GetIsGlobalSEO get QorSEOSetting's isGlobal
func (s QorSEOSetting) GetIsGlobalSEO() bool {
	return s.IsGlobalSEO
}

// SetIsGlobalSEO set QorSeoSetting's isGlobal
func (s *QorSEOSetting) SetIsGlobalSEO(isGlobal bool) {
	s.IsGlobalSEO = isGlobal
}

// GetGlobalSetting get QorSeoSetting's globalSetting
func (s QorSEOSetting) GetGlobalSetting() map[string]string {
	return s.Setting.GlobalSetting
}

// SetGlobalSetting set QorSeoSetting's globalSetting
func (s *QorSEOSetting) SetGlobalSetting(globalSetting map[string]string) {
	s.Setting.GlobalSetting = globalSetting
}

// GetTitle get Setting's title
func (s QorSEOSetting) GetTitle() string {
	return s.Setting.Title
}

// GetDescription get Setting's description
func (s QorSEOSetting) GetDescription() string {
	return s.Setting.Description
}

// GetKeywords get Setting's keywords
func (s QorSEOSetting) GetKeywords() string {
	return s.Setting.Keywords
}

// SetCollection set Setting's collection
func (s *QorSEOSetting) SetCollection(collection *Collection) {
	s.collection = collection
}

// GetSEO get Setting's SEO configure
func (s QorSEOSetting) GetSEO() *SEO {
	return s.collection.GetSEO(s.Name)
}

// Scan scan value from database into struct
func (setting *Setting) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		json.Unmarshal(bytes, setting)
	} else if str, ok := value.(string); ok {
		json.Unmarshal([]byte(str), setting)
	} else if strs, ok := value.([]string); ok {
		for _, str := range strs {
			json.Unmarshal([]byte(str), setting)
		}
	}
	return nil
}

// Value get value from struct, and save into database
func (setting Setting) Value() (driver.Value, error) {
	result, err := json.Marshal(setting)
	return string(result), err
}

// String return formated seo setting
func (setting Setting) String() string {
	basicMeta := fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">\n<meta name=\"keywords\" content=\"%s\"/>", setting.Title, setting.Description, setting.Keywords)

	if setting.OpenGraphURL != "" {
		basicMeta += fmt.Sprintf("<meta property=\"og:title\" content=\"%v\"/>", setting.Title)
		basicMeta += fmt.Sprintf("<meta property=\"og:url\" content=\"%v\"/>", setting.OpenGraphURL)
	}

	if setting.OpenGraphType != "" {
		basicMeta += fmt.Sprintf("<meta property=\"og:type\" content=\"%v\"/>", setting.OpenGraphType)
	}

	if len(setting.OpenGraphImage.Files) > 0 {
		basicMeta += fmt.Sprintf("<meta property=\"og:image\" content=\"%v\"/>", setting.OpenGraphImage.URL())
	}

	for _, metavalue := range setting.OpenGraphMetadata {
		basicMeta += fmt.Sprintf("<meta property=\"%v\" content=\"%v\"/>", metavalue.Property, metavalue.Content)
	}

	return basicMeta
}

// ConfigureQorMetaBeforeInitialize configure SEO setting for qor admin
func (setting Setting) ConfigureQorMetaBeforeInitialize(meta resource.Metaor) {
	if meta, ok := meta.(*admin.Meta); ok {
		meta.Type = "seo"
		if res, ok := meta.GetBaseResource().(*admin.Resource); ok {
			res.UseTheme("seo_meta")
		}
	}
}

// ConfigureQorResource configure resource for seo setting
func (setting Setting) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		res.Meta(&admin.Meta{Name: "Title", Label: "HTML Title"})
		res.Meta(&admin.Meta{Name: "Description", Label: "Meta Description"})
		res.Meta(&admin.Meta{Name: "Keywords", Label: "Meta Keywords"})
		res.Meta(&admin.Meta{Name: "Type", Type: "hidden"})
		res.Meta(&admin.Meta{Name: "EnabledCustomize", Type: "hidden", Valuer: func(interface{}, *qor.Context) interface{} { return "" }})
		res.Meta(&admin.Meta{Name: "OpenGraphImage", Config: &media_library.MediaBoxConfig{
			Max: 1,
		}})

		metadataResource := res.Meta(&admin.Meta{Name: "OpenGraphMetadata"}).Resource
		metadataResource.NewAttrs(&admin.Section{Rows: [][]string{{"Property", "Content"}}})
		metadataResource.EditAttrs(&admin.Section{Rows: [][]string{{"Property", "Content"}}})

		res.EditAttrs("Title", "Description", "Keywords", "Type", "OpenGraphURL", "OpenGraphType", "OpenGraphImage", "OpenGraphMetadata", "EnabledCustomize")
	}
}
