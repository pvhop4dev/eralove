package i18n

import (
	"encoding/json"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

// I18n handles internationalization
type I18n struct {
	bundle   *i18n.Bundle
	logger   *zap.Logger
	fallback language.Tag
}

// NewI18n creates a new i18n instance
func NewI18n(logger *zap.Logger) *I18n {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	return &I18n{
		bundle:   bundle,
		logger:   logger,
		fallback: language.English,
	}
}

// LoadMessages loads translation messages from files
func (i *I18n) LoadMessages(messagesDir string) error {
	// Load English messages
	enFile := filepath.Join(messagesDir, "en.json")
	if _, err := i.bundle.LoadMessageFile(enFile); err != nil {
		i.logger.Warn("Failed to load English messages", zap.Error(err))
	}

	// Load Spanish messages
	esFile := filepath.Join(messagesDir, "es.json")
	if _, err := i.bundle.LoadMessageFile(esFile); err != nil {
		i.logger.Warn("Failed to load Spanish messages", zap.Error(err))
	}

	// Load French messages
	frFile := filepath.Join(messagesDir, "fr.json")
	if _, err := i.bundle.LoadMessageFile(frFile); err != nil {
		i.logger.Warn("Failed to load French messages", zap.Error(err))
	}

	i.logger.Info("Translation messages loaded")
	return nil
}

// GetLocalizer returns a localizer for the given language
func (i *I18n) GetLocalizer(lang string) *i18n.Localizer {
	var acceptLanguages []string

	// Parse the requested language
	if tag, err := language.Parse(lang); err == nil {
		acceptLanguages = append(acceptLanguages, tag.String())
	}

	// Add fallback language
	acceptLanguages = append(acceptLanguages, i.fallback.String())

	return i18n.NewLocalizer(i.bundle, acceptLanguages...)
}

// Translate translates a message key with optional template data
func (i *I18n) Translate(lang, messageID string, templateData map[string]interface{}) string {
	// Parse Accept-Language header if it contains multiple languages
	parsedLang := i.ParseAcceptLanguage(lang)
	
	localizer := i.GetLocalizer(parsedLang)

	config := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}

	translation, err := localizer.Localize(config)
	if err != nil {
		i.logger.Warn("Translation not found",
			zap.String("messageID", messageID),
			zap.String("lang", lang),
			zap.String("parsed_lang", parsedLang),
			zap.Error(err))
		return messageID // Return the message ID as fallback
	}

	return translation
}

// GetSupportedLanguages returns list of supported languages
func (i *I18n) GetSupportedLanguages() []string {
	return []string{"en", "es", "fr"}
}

// IsLanguageSupported checks if a language is supported
func (i *I18n) IsLanguageSupported(lang string) bool {
	supported := i.GetSupportedLanguages()
	for _, supportedLang := range supported {
		if supportedLang == lang {
			return true
		}
	}
	return false
}

// ParseAcceptLanguage parses Accept-Language header and returns best match
func (i *I18n) ParseAcceptLanguage(acceptLang string) string {
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil {
		return "en" // Default to English
	}

	supported := []language.Tag{
		language.English,
		language.Spanish,
		language.French,
	}

	matcher := language.NewMatcher(supported)
	tag, _, _ := matcher.Match(tags...)

	switch tag {
	case language.Spanish:
		return "es"
	case language.French:
		return "fr"
	default:
		return "en"
	}
}
