package concourse

func SanitizedSource(source Source) map[string]string {
	s := make(map[string]string)

	if source.APIToken != "" {
		s[source.APIToken] = "***REDACTED-PIVNET_API_TOKEN***"
	}
	if source.AccessKeyID != "" {
		s[source.AccessKeyID] = "***REDACTED-AWS_ACCESS_KEY_ID***"
	}
	if source.SecretAccessKey != "" {
		s[source.SecretAccessKey] = "***REDACTED-AWS_SECRET_ACCESS_KEY***"
	}
	if source.Username != "" {
		s[source.Username] = "***REDACTED-UAA_USERNAME***"
	}
	if source.Password != "" {
		s[source.Password] = "***REDACTED-UAA_PASSWORD***"
	}

	return s
}
