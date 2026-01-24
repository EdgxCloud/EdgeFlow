package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============ MySQL Tests ============

func TestNewMySQLExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"host":     "localhost",
				"database": "test_db",
				"username": "root",
				"password": "secret",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: map[string]interface{}{
				"database": "test_db",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: map[string]interface{}{
				"host": "localhost",
			},
			wantErr: true,
		},
		{
			name: "default port",
			config: map[string]interface{}{
				"host":     "localhost",
				"database": "test_db",
			},
			wantErr: false,
		},
		{
			name: "custom port",
			config: map[string]interface{}{
				"host":     "localhost",
				"port":     3307,
				"database": "test_db",
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"host":     "db.example.com",
				"port":     3306,
				"database": "production",
				"username": "app_user",
				"password": "secure_password",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewMySQLExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestMySQLConfig_DefaultPort(t *testing.T) {
	executor, err := NewMySQLExecutor(map[string]interface{}{
		"host":     "localhost",
		"database": "test_db",
	})
	require.NoError(t, err)

	mysqlExecutor := executor.(*MySQLExecutor)
	assert.Equal(t, 3306, mysqlExecutor.config.Port)
}

func TestMySQLExecutor_Cleanup(t *testing.T) {
	executor, err := NewMySQLExecutor(map[string]interface{}{
		"host":     "localhost",
		"database": "test_db",
	})
	require.NoError(t, err)

	// Cleanup should not error without connection
	err = executor.Cleanup()
	assert.NoError(t, err)
}

// ============ PostgreSQL Tests ============

func TestNewPostgreSQLExecutor_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "missing host",
			config: map[string]interface{}{
				"database": "test_db",
				"username": "postgres",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: map[string]interface{}{
				"host":     "localhost",
				"username": "postgres",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: map[string]interface{}{
				"host":     "localhost",
				"database": "test_db",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPostgreSQLExecutor(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// ============ Redis Tests ============

func TestNewRedisExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "default config",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "with host and port",
			config: map[string]interface{}{
				"host": "redis.example.com",
				"port": 6380,
			},
			wantErr: false,
		},
		{
			name: "with password",
			config: map[string]interface{}{
				"host":     "redis.example.com",
				"password": "secret",
			},
			wantErr: false,
		},
		{
			name: "with database number",
			config: map[string]interface{}{
				"host": "localhost",
				"db":   1,
			},
			wantErr: false,
		},
		{
			name: "full config",
			config: map[string]interface{}{
				"host":     "redis.example.com",
				"port":     6379,
				"password": "secret",
				"db":       2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewRedisExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestRedisConfig_Defaults(t *testing.T) {
	executor, err := NewRedisExecutor(map[string]interface{}{})
	require.NoError(t, err)

	redisExecutor := executor.(*RedisExecutor)
	assert.Equal(t, "localhost", redisExecutor.config.Host)
	assert.Equal(t, 6379, redisExecutor.config.Port)
	assert.Equal(t, 0, redisExecutor.config.DB)
}

func TestRedisExecutor_Cleanup(t *testing.T) {
	executor, err := NewRedisExecutor(map[string]interface{}{})
	require.NoError(t, err)

	// Cleanup should not error without connection
	err = executor.Cleanup()
	assert.NoError(t, err)
}

// ============ MongoDB Tests ============

func TestNewMongoDBExecutor_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "missing uri",
			config: map[string]interface{}{
				"database": "test_db",
			},
			wantErr: true,
		},
		{
			name: "missing database",
			config: map[string]interface{}{
				"uri": "mongodb://localhost:27017",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMongoDBExecutor(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// ============ Config Parsing Tests ============

func TestMySQLConfig_Parsing(t *testing.T) {
	executor, err := NewMySQLExecutor(map[string]interface{}{
		"host":     "db.example.com",
		"port":     3307,
		"database": "mydb",
		"username": "user",
		"password": "pass",
	})
	require.NoError(t, err)

	mysqlExecutor := executor.(*MySQLExecutor)
	assert.Equal(t, "db.example.com", mysqlExecutor.config.Host)
	assert.Equal(t, 3307, mysqlExecutor.config.Port)
	assert.Equal(t, "mydb", mysqlExecutor.config.Database)
	assert.Equal(t, "user", mysqlExecutor.config.Username)
	assert.Equal(t, "pass", mysqlExecutor.config.Password)
}

func TestRedisConfig_Parsing(t *testing.T) {
	executor, err := NewRedisExecutor(map[string]interface{}{
		"host":     "redis.example.com",
		"port":     6380,
		"password": "secret",
		"db":       5,
	})
	require.NoError(t, err)

	redisExecutor := executor.(*RedisExecutor)
	assert.Equal(t, "redis.example.com", redisExecutor.config.Host)
	assert.Equal(t, 6380, redisExecutor.config.Port)
	assert.Equal(t, "secret", redisExecutor.config.Password)
	assert.Equal(t, 5, redisExecutor.config.DB)
}
