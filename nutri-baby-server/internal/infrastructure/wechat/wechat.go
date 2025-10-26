package wechat

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/config"
)

// Client 微信客户端封装
type Client struct {
	wechat      *wechat.Wechat
	miniProgram *miniprogram.MiniProgram
}

// NewClient 创建微信客户端
func NewClient(cfg *config.Config, redisClient *redis.Client) *Client {
	// 创建 Redis 缓存适配器
	// Host 需要包含端口号，格式为 "host:port"
	redisHost := cfg.Redis.Host
	if cfg.Redis.Port != 0 {
		redisHost = fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	}

	redisCache := cache.NewRedis(context.Background(), &cache.RedisOpts{
		Host:     redisHost,
		Password: cfg.Redis.Password,
		Database: cfg.Redis.DB,
	})

	// 创建微信实例
	wc := wechat.NewWechat()

	// 配置小程序
	miniCfg := &miniConfig.Config{
		AppID:     cfg.Wechat.AppID,
		AppSecret: cfg.Wechat.AppSecret,
		Cache:     redisCache,
	}

	// 获取小程序实例
	mini := wc.GetMiniProgram(miniCfg)

	return &Client{
		wechat:      wc,
		miniProgram: mini,
	}
}

// GetMiniProgram 获取小程序实例
func (c *Client) GetMiniProgram() *miniprogram.MiniProgram {
	return c.miniProgram
}
