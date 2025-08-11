package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 获取数据库连接信息
	fmt.Print("请输入数据库连接信息:\n")
	fmt.Print("Host (默认 localhost): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}

	fmt.Print("Port (默认 5432): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)
	if port == "" {
		port = "5432"
	}

	fmt.Print("数据库名: ")
	dbname, _ := reader.ReadString('\n')
	dbname = strings.TrimSpace(dbname)
	if dbname == "" {
		log.Fatal("数据库名不能为空")
	}

	fmt.Print("数据库用户名: ")
	dbUser, _ := reader.ReadString('\n')
	dbUser = strings.TrimSpace(dbUser)
	if dbUser == "" {
		log.Fatal("数据库用户名不能为空")
	}

	fmt.Print("数据库密码: ")
	dbPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("读取数据库密码失败:", err)
	}
	fmt.Println() // 换行
	dbPassword := strings.TrimSpace(string(dbPasswordBytes))

	// 获取要重置密码的用户名
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

	// 连接数据库
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, dbUser, dbPassword, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	err = db.Ping()
	if err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Printf("✅ 数据库连接成功！\n")

	// 检查用户是否存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	if err != nil {
		log.Fatal("查询用户失败:", err)
	}

	if count == 0 {
		log.Fatalf("用户 '%s' 不存在", username)
	}

	// 更新密码
	_, err = db.Exec("UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE username = $2",
		string(hashedPassword), username)
	if err != nil {
		log.Fatal("更新密码失败:", err)
	}

	fmt.Printf("✅ 用户 '%s' 的密码已成功重置！\n", username)

	// 验证新密码
	fmt.Printf("验证新密码...\n")
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		log.Fatal("密码验证失败:", err)
	}
	fmt.Printf("✅ 密码验证成功！\n")
}
