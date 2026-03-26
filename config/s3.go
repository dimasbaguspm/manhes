package config

import "os"

type S3Config struct {
	Endpoint       string
	Bucket         string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	PublicEndpoint string
}

func loadS3Config() S3Config {
	return S3Config{
		Endpoint:       envStr("S3_ENDPOINT", "localhost:9000"),
		Bucket:         envStr("S3_BUCKET", "manga"),
		AccessKey:      os.Getenv("S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("S3_SECRET_KEY"),
		UseSSL:         envBool("S3_USE_SSL", false),
		PublicEndpoint: os.Getenv("S3_PUBLIC_ENDPOINT"),
	}
}
