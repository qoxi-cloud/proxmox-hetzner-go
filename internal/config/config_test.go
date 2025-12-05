package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSystemConfig_SensitiveFieldsOmittedFromYAML(t *testing.T) {
	tests := []struct {
		name             string
		cfg              SystemConfig
		shouldNotContain []string
		shouldContain    []string
	}{
		{
			name: "standard config with all fields",
			cfg: SystemConfig{
				Hostname:     "pve-server",
				DomainSuffix: "local",
				Timezone:     "Europe/Kyiv",
				Email:        "admin@example.com",
				RootPassword: "secret-password",
				SSHPublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIG...",
			},
			shouldNotContain: []string{"secret-password", "ssh-ed25519", "root_password", "ssh_public_key"},
			shouldContain:    []string{"hostname: pve-server", "domain_suffix: local", "timezone: Europe/Kyiv", "email: admin@example.com"},
		},
		{
			name: "config with special characters in sensitive fields",
			cfg: SystemConfig{
				Hostname:     "test-host",
				DomainSuffix: "example.com",
				Timezone:     "UTC",
				Email:        "test@test.com",
				RootPassword: "p@ssw0rd!#$%",
				SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB...",
			},
			shouldNotContain: []string{"p@ssw0rd!#$%", "ssh-rsa AAAAB3", "root_password", "ssh_public_key"},
			shouldContain:    []string{"hostname: test-host", "domain_suffix: example.com"},
		},
		{
			name: "config with empty sensitive fields",
			cfg: SystemConfig{
				Hostname:     "empty-secrets",
				DomainSuffix: "local",
				Timezone:     "UTC",
				Email:        "admin@local",
				RootPassword: "",
				SSHPublicKey: "",
			},
			shouldNotContain: []string{"root_password", "ssh_public_key"},
			shouldContain:    []string{"hostname: empty-secrets"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			yamlStr := string(data)

			for _, str := range tt.shouldNotContain {
				assert.NotContains(t, yamlStr, str)
			}
			for _, str := range tt.shouldContain {
				assert.Contains(t, yamlStr, str)
			}
		})
	}
}

func TestSystemConfig_StandardFieldsSerializeCorrectly(t *testing.T) {
	tests := []struct {
		name string
		cfg  SystemConfig
	}{
		{
			name: "standard values",
			cfg: SystemConfig{
				Hostname:     "test-server",
				DomainSuffix: "example.com",
				Timezone:     "UTC",
				Email:        "test@example.com",
			},
		},
		{
			name: "empty strings",
			cfg: SystemConfig{
				Hostname:     "",
				DomainSuffix: "",
				Timezone:     "",
				Email:        "",
			},
		},
		{
			name: "special characters",
			cfg: SystemConfig{
				Hostname:     "pve-server-01",
				DomainSuffix: "sub.domain.example.com",
				Timezone:     "America/New_York",
				Email:        "admin+alerts@example.com",
			},
		},
		{
			name: "unicode characters",
			cfg: SystemConfig{
				Hostname:     "server",
				DomainSuffix: "example.com",
				Timezone:     "Europe/Kyiv",
				Email:        "user@example.com",
			},
		},
		{
			name: "long values",
			cfg: SystemConfig{
				Hostname:     "very-long-hostname-that-is-still-valid",
				DomainSuffix: "subdomain.another.level.example.com",
				Timezone:     "America/Argentina/Buenos_Aires",
				Email:        "very.long.email.address@subdomain.example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.cfg)
			require.NoError(t, err)

			var result SystemConfig
			err = yaml.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.cfg.Hostname, result.Hostname)
			assert.Equal(t, tt.cfg.DomainSuffix, result.DomainSuffix)
			assert.Equal(t, tt.cfg.Timezone, result.Timezone)
			assert.Equal(t, tt.cfg.Email, result.Email)
		})
	}
}

func TestSystemConfig_RoundTripMarshalUnmarshal(t *testing.T) {
	original := SystemConfig{
		Hostname:     "production-pve",
		DomainSuffix: "prod.example.com",
		Timezone:     "America/New_York",
		Email:        "ops@company.com",
		RootPassword: "super-secret",
		SSHPublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&original)
	require.NoError(t, err)

	// Unmarshal back
	var restored SystemConfig
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Non-sensitive fields should be restored
	assert.Equal(t, original.Hostname, restored.Hostname)
	assert.Equal(t, original.DomainSuffix, restored.DomainSuffix)
	assert.Equal(t, original.Timezone, restored.Timezone)
	assert.Equal(t, original.Email, restored.Email)

	// Sensitive fields should be empty after round-trip (they were excluded)
	assert.Empty(t, restored.RootPassword)
	assert.Empty(t, restored.SSHPublicKey)
}

func TestSystemConfig_EnvironmentVariableTagsPresent(t *testing.T) {
	expectedEnvTags := map[string]string{
		"Hostname":     "PVE_HOSTNAME",
		"DomainSuffix": "PVE_DOMAIN_SUFFIX",
		"Timezone":     "PVE_TIMEZONE",
		"Email":        "PVE_EMAIL",
		"RootPassword": "PVE_ROOT_PASSWORD",
		"SSHPublicKey": "PVE_SSH_PUBLIC_KEY",
	}

	cfgType := reflect.TypeOf(SystemConfig{})

	for fieldName, expectedTag := range expectedEnvTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		envTag := field.Tag.Get("env")
		assert.Equal(t, expectedTag, envTag, "env tag mismatch for field %s", fieldName)
	}
}

func TestSystemConfig_YAMLTagsPresent(t *testing.T) {
	expectedYAMLTags := map[string]string{
		"Hostname":     "hostname",
		"DomainSuffix": "domain_suffix",
		"Timezone":     "timezone",
		"Email":        "email",
		"RootPassword": "-",
		"SSHPublicKey": "-",
	}

	cfgType := reflect.TypeOf(SystemConfig{})

	for fieldName, expectedTag := range expectedYAMLTags {
		field, found := cfgType.FieldByName(fieldName)
		require.True(t, found, "field %s not found", fieldName)

		yamlTag := field.Tag.Get("yaml")
		assert.Equal(t, expectedTag, yamlTag, "yaml tag mismatch for field %s", fieldName)
	}
}

func TestSystemConfig_AllFieldsExist(t *testing.T) {
	requiredFields := []string{
		"Hostname",
		"DomainSuffix",
		"Timezone",
		"Email",
		"RootPassword",
		"SSHPublicKey",
	}

	cfgType := reflect.TypeOf(SystemConfig{})

	assert.Equal(t, len(requiredFields), cfgType.NumField(), "unexpected number of fields")

	for _, fieldName := range requiredFields {
		field, found := cfgType.FieldByName(fieldName)
		assert.True(t, found, "required field %s not found", fieldName)
		assert.Equal(t, "string", field.Type.Kind().String(), "field %s should be string type", fieldName)
	}
}
