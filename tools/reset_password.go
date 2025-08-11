package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	
	// 获取用户名
	fmt.Print("请输入要重置密码的用户名: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	
	if username == "" {
		log.Fatal("用户名不能为空")
	}
	
	// 获取新密码
	fmt.Print("请输入新密码: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("读取密码失败:", err)
	}
	fmt.Println() // 换行
	
	password := strings.TrimSpace(string(passwordBytes))
	if password == "" {
		log.Fatal("密码不能为空")
	}
	
	// 确认密码
	fmt.Print("请再次输入密码确认: ")
	confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("读取确认密码失败:", err)
	}
	fmt.Println() // 换行
	
	confirmPassword := strings.TrimSpace(string(confirmPasswordBytes))
	if password != confirmPassword {
		log.Fatal("两次输入的密码不一致")
	}
	
	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("生成密码哈希失败:", err)
	}
	
	fmt.Printf("\n用户 '%s' 的新密码哈希值:\n", username)
	fmt.Printf("%s\n\n", string(hashedPassword))
	
	// 生成 SQL 更新语句
	fmt.Printf("请执行以下 SQL 语句来更新数据库中的密码:\n")
	fmt.Printf("UPDATE users SET password = '%s' WHERE username = '%s';\n\n", string(hashedPassword), username)
	
	// 验证新密码
	fmt.Printf("验证新密码...\n")
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		log.Fatal("密码验证失败:", err)
	}
	fmt.Printf("✅ 密码验证成功！\n")
}
