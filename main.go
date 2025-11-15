package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoManager struct {
	client        *mongo.Client
	ctx           context.Context
	connectionURL string
	host          string
	port          string
}

func NewMongoManager(connectionURL string) (*MongoManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(connectionURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar: %w", err)
	}

	// Verifica a conexão
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer ping: %w", err)
	}

	fmt.Println("✓ Conectado ao MongoDB com sucesso!")

	// Extrai host e port da URL de conexão
	host, port := extractHostAndPort(connectionURL)

	return &MongoManager{
		client:        client,
		ctx:           context.Background(),
		connectionURL: connectionURL,
		host:          host,
		port:          port,
	}, nil
}

// extractHostAndPort extrai o host e porta da URL de conexão MongoDB
func extractHostAndPort(connectionURL string) (string, string) {
	// Remove o prefixo mongodb:// ou mongodb+srv://
	urlStr := connectionURL
	if strings.HasPrefix(urlStr, "mongodb+srv://") {
		urlStr = strings.TrimPrefix(urlStr, "mongodb+srv://")
		// Para mongodb+srv, não temos porta explícita
		parts := strings.Split(urlStr, "@")
		if len(parts) == 2 {
			hostPart := strings.Split(parts[1], "/")[0]
			hostPart = strings.Split(hostPart, "?")[0]
			return hostPart, ""
		}
		return "localhost", "27017"
	}

	if strings.HasPrefix(urlStr, "mongodb://") {
		urlStr = strings.TrimPrefix(urlStr, "mongodb://")
	}

	// Remove credenciais se existirem
	parts := strings.Split(urlStr, "@")
	var hostPart string
	if len(parts) == 2 {
		hostPart = parts[1]
	} else {
		hostPart = parts[0]
	}

	// Remove database e query params
	hostPart = strings.Split(hostPart, "/")[0]
	hostPart = strings.Split(hostPart, "?")[0]

	// Separa host e port
	hostPort := strings.Split(hostPart, ":")
	if len(hostPort) == 2 {
		return hostPort[0], hostPort[1]
	}
	return hostPort[0], "27017"
}

func (m *MongoManager) CreateDatabase(dbName string) error {
	// No MongoDB, databases são criadas automaticamente quando você insere dados
	// Vamos criar uma collection temporária para garantir que a DB existe
	db := m.client.Database(dbName)

	// Cria uma collection temporária
	err := db.CreateCollection(m.ctx, "_init")
	if err != nil {
		return fmt.Errorf("erro ao criar database: %w", err)
	}

	fmt.Printf("✓ Database '%s' criada com sucesso!\n", dbName)
	return nil
}

func (m *MongoManager) CreateUser(dbName, username, password string, roles []string) error {
	// Verifica se a database existe antes de criar o usuário
	exists, err := m.DatabaseExists(dbName)
	if err != nil {
		return fmt.Errorf("erro ao verificar database: %w", err)
	}

	if !exists {
		return fmt.Errorf("a database '%s' não existe. Crie a database primeiro antes de criar usuários", dbName)
	}

	db := m.client.Database(dbName)

	// Define as roles do usuário
	userRoles := make([]bson.M, 0)
	for _, role := range roles {
		userRoles = append(userRoles, bson.M{
			"role": role,
			"db":   dbName,
		})
	}

	// Comando para criar usuário
	command := bson.D{
		{Key: "createUser", Value: username},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: userRoles},
	}

	var result bson.M
	err = db.RunCommand(m.ctx, command).Decode(&result)
	if err != nil {
		return fmt.Errorf("erro ao criar usuário: %w", err)
	}

	fmt.Printf("✓ Usuário '%s' criado com sucesso na database '%s'!\n", username, dbName)
	fmt.Printf("   Roles atribuídas: %v\n", roles)

	// Gera a connection string com escape de caracteres especiais
	connectionString := m.generateConnectionString(username, password, dbName)
	fmt.Printf("\n📋 Connection String:\n%s\n", connectionString)

	return nil
}

// generateConnectionString gera uma URL de conexão válida com escape de caracteres especiais
func (m *MongoManager) generateConnectionString(username, password, dbName string) string {
	// Escapa caracteres especiais no username e password
	escapedUsername := url.QueryEscape(username)
	escapedPassword := url.QueryEscape(password)

	// Verifica se é mongodb+srv
	if strings.Contains(m.connectionURL, "mongodb+srv://") {
		return fmt.Sprintf("mongodb+srv://%s:%s@%s/%s",
			escapedUsername, escapedPassword, m.host, dbName)
	}

	// MongoDB padrão com host:port
	if m.port != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			escapedUsername, escapedPassword, m.host, m.port, dbName)
	}

	// Fallback sem porta
	return fmt.Sprintf("mongodb://%s:%s@%s/%s",
		escapedUsername, escapedPassword, m.host, dbName)
}

// DatabaseExists verifica se uma database existe
func (m *MongoManager) DatabaseExists(dbName string) (bool, error) {
	databases, err := m.client.ListDatabaseNames(m.ctx, bson.M{})
	if err != nil {
		return false, err
	}

	for _, db := range databases {
		if db == dbName {
			return true, nil
		}
	}
	return false, nil
}

func (m *MongoManager) ListDatabases() error {
	databases, err := m.client.ListDatabaseNames(m.ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("erro ao listar databases: %w", err)
	}

	fmt.Println("\nDatabases disponíveis:")
	for _, db := range databases {
		fmt.Printf("  - %s\n", db)
	}
	return nil
}

func (m *MongoManager) Close() {
	if m.client != nil {
		m.client.Disconnect(m.ctx)
	}
}

func main() {
	fmt.Println("=== Go MongoDB Database & User Manager ===\n")

	// Solicita a URL de conexão de forma segura (não fica no histórico do terminal)
	fmt.Println("Digite a URL de conexão do MongoDB:")
	fmt.Println("Exemplos:")
	fmt.Println("  mongodb://localhost:27017")
	fmt.Println("  mongodb://admin:senha@localhost:27017")
	fmt.Println("  mongodb+srv://user:pass@cluster.mongodb.net/")
	fmt.Print("\nURL de conexão: ")

	var connectionURL string
	fmt.Scanln(&connectionURL)

	if connectionURL == "" {
		log.Fatal("✗ URL de conexão não pode ser vazia")
	}

	fmt.Println("\nConectando ao MongoDB...")
	manager, err := NewMongoManager(connectionURL)
	if err != nil {
		log.Fatalf("✗ Erro ao inicializar: %v", err)
	}
	defer manager.Close()

	// Menu interativo
	for {
		fmt.Println("\n=== Menu ===")
		fmt.Println("1. Criar Database")
		fmt.Println("2. Criar Usuário")
		fmt.Println("3. Listar Databases")
		fmt.Println("4. Sair")
		fmt.Print("\nEscolha uma opção: ")

		var opcao int
		fmt.Scanln(&opcao)

		switch opcao {
		case 1:
			fmt.Print("Nome da database: ")
			var dbName string
			fmt.Scanln(&dbName)

			if dbName == "" {
				fmt.Println("✗ Nome da database não pode ser vazio")
				continue
			}

			err := manager.CreateDatabase(dbName)
			if err != nil {
				fmt.Printf("✗ Erro: %v\n", err)
			}

		case 2:
			fmt.Print("Nome da database: ")
			var dbName string
			fmt.Scanln(&dbName)

			if dbName == "" {
				fmt.Println("✗ Nome da database não pode ser vazio")
				continue
			}

			// Verifica se a database existe
			exists, err := manager.DatabaseExists(dbName)
			if err != nil {
				fmt.Printf("✗ Erro ao verificar database: %v\n", err)
				continue
			}

			if !exists {
				fmt.Printf("⚠️  A database '%s' não existe.\n", dbName)
				fmt.Print("Deseja criar a database agora? (S/n): ")
				var resposta string
				fmt.Scanln(&resposta)

				// Se vazio ou S/s, cria a database
				if resposta == "" || resposta == "S" || resposta == "s" {
					err := manager.CreateDatabase(dbName)
					if err != nil {
						fmt.Printf("✗ Erro ao criar database: %v\n", err)
						continue
					}
				} else {
					fmt.Println("✗ Operação cancelada. Crie a database primeiro.")
					continue
				}
			}

			fmt.Print("Nome do usuário: ")
			var username string
			fmt.Scanln(&username)

			fmt.Print("Senha: ")
			var password string
			fmt.Scanln(&password)

			fmt.Println("\nRoles disponíveis: read, readWrite, dbAdmin, userAdmin, dbOwner")
			fmt.Print("Roles (separadas por vírgula, ex: read,readWrite): ")
			var rolesInput string
			fmt.Scanln(&rolesInput)

			// Se não especificar roles, usa readWrite como padrão
			roles := []string{"readWrite"}
			if rolesInput != "" {
				roles = []string{}
				for i := 0; i < len(rolesInput); i++ {
					var role string
					for i < len(rolesInput) && rolesInput[i] != ',' {
						role += string(rolesInput[i])
						i++
					}
					if role != "" {
						roles = append(roles, role)
					}
				}
			}

			if dbName == "" || username == "" || password == "" {
				fmt.Println("✗ Todos os campos são obrigatórios")
				continue
			}

			err = manager.CreateUser(dbName, username, password, roles)
			if err != nil {
				fmt.Printf("✗ Erro: %v\n", err)
			}

		case 3:
			err := manager.ListDatabases()
			if err != nil {
				fmt.Printf("✗ Erro: %v\n", err)
			}

		case 4:
			fmt.Println("Até logo!")
			return

		default:
			fmt.Println("✗ Opção inválida")
		}
	}
}
