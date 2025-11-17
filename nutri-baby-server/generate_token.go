package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// 命令行参数
	openid := flag.String("openid", "om8hB12mqHOp1BiTf3KZ_ew8eWH4", "用户openid")
	secret := flag.String("secret", "your-secret-key-change-in-production", "JWT密钥（需要与config.yaml中的一致）")
	expireHours := flag.Int("expire", 72, "token过期时间（小时）")
	flag.Parse()

	// 创建JWT令牌
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   *openid,
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(*expireHours) * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(*secret))
	if err != nil {
		fmt.Printf("生成token失败: %v\n", err)
		return
	}

	fmt.Printf("✅ Token生成成功:\n\n")
	fmt.Printf("Bearer Token:\n%s\n\n", tokenString)
	fmt.Printf("Token信息:\n")
	fmt.Printf("  - OpenID: %s\n", *openid)
	fmt.Printf("  - Secret: %s\n", *secret)
	fmt.Printf("  - 过期时间: %d小时\n", *expireHours)
	fmt.Printf("  - 生成时间: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("  - 过期时间: %s\n\n", now.Add(time.Duration(*expireHours)*time.Hour).Format("2006-01-02 15:04:05"))

	fmt.Printf("在APIfox中使用:\n")
	fmt.Printf("Authorization: Bearer %s\n", tokenString)
}
