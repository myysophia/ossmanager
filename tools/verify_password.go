package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// demo 用户的密码哈希
	hash := "$2a$10$ZT.krA9FFglHDZs08U42hOTmrrY9IOCWU1NrN266UfEYUF9moQisq"
	
	// 常见密码列表
	commonPasswords := []string{
		"demo",
		"Demo",
		"DEMO",
		"demo123",
		"Demo123",
		"DEMO123",
		"password",
		"Password",
		"PASSWORD",
		"123456",
		"admin",
		"admin123",
		"test",
		"test123",
		"user",
		"user123",
		"demo@123",
		"demo_123",
		"demouser",
		"demopwd",
		"12345678",
		"qwerty",
		"abc123",
		"000000",
		"111111",
		"888888",
		"666666",
		"123123",
		"demo2023",
		"demo2024",
		"demo2025",
	}
	
	fmt.Printf("正在尝试验证 demo 用户的密码...\n")
	fmt.Printf("密码哈希: %s\n\n", hash)
	
	for i, password := range commonPasswords {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err == nil {
			fmt.Printf("🎉 找到密码了！\n")
			fmt.Printf("用户名: demo\n")
			fmt.Printf("密码: %s\n", password)
			return
		}
		fmt.Printf("尝试 %d/%d: %s ... ❌\n", i+1, len(commonPasswords), password)
	}
	
	fmt.Printf("\n❌ 未能在常见密码列表中找到匹配的密码。\n")
	fmt.Printf("建议使用密码重置工具重置密码。\n")
}
