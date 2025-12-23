package redis

import "log"

func InitRedis() (Client, error) {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load Redis configuration: %v", err)
		return nil, err
	}

	client, err := newRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis connection: %v", err)
		return nil, err
	}

	return client, nil
}
