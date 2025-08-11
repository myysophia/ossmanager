package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 为 demo 用户生成几个常用的新密码选项
	passwords := []string{
		"demo123",
		"Demo@123",
		"demo2025",
		"123456",
		"password",
	}
	
	fmt.Printf("为 demo 用户生成新密码哈希值：\n\n")
	
	for i, password := range passwords {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("生成密码 %s 失败: %v\n", password, err)
			continue
		}
		
		fmt.Printf("选项 %d:\n", i+1)
		fmt.Printf("  密码: %s\n", password)
		fmt.Printf("  哈希: %s\n", string(hashedPassword))
		fmt.Printf("  SQL: UPDATE users SET password = '%s' WHERE username = 'demo';\n\n", string(hashedPassword))
	}
	
	fmt.Printf("🔧 使用方法:\n")
	fmt.Printf("1. 选择一个你喜欢的密码选项\n")
	fmt.Printf("2. 复制对应的 SQL 语句\n")
	fmt.Printf("3. 连接到你的数据库执行该 SQL 语句\n")
	fmt.Printf("4. 使用新密码登录\n\n")
	
	fmt.Printf("💡 如果你有数据库连接，也可以使用: go run reset_password_db.go\n")
}
