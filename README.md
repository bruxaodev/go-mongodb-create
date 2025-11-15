# Go MongoDB Database & User Manager

Ferramenta CLI em Go para gerenciar databases e usuários no MongoDB.

## Funcionalidades

- ✅ Conectar ao MongoDB usando URL de conexão
- ✅ **Entrada segura da URL** - Não fica salva no histórico do terminal
- ✅ Criar databases
- ✅ Criar usuários com permissões específicas para databases
- ✅ Listar databases disponíveis
- ✅ Interface interativa via menu
- ✅ **Geração automática de connection string** com:
  - Escape automático de caracteres especiais na senha
  - Host e porta extraídos da URL original
  - Suporte para MongoDB padrão e MongoDB Atlas (mongodb+srv)

## Instalação

```bash
go mod download
go build -o go-db-create
```

## Uso

### Iniciar o Programa

Simplesmente execute o programa:

```bash
./go-db-create
```

O programa irá solicitar a URL de conexão de forma segura (não fica no histórico do terminal):

```
=== Go MongoDB Database & User Manager ===

Digite a URL de conexão do MongoDB:
Exemplos:
  mongodb://localhost:27017
  mongodb://admin:senha@localhost:27017
  mongodb+srv://user:pass@cluster.mongodb.net/

URL de conexão: mongodb://admin:senha@localhost:27017

Conectando ao MongoDB...
✓ Conectado ao MongoDB com sucesso!
```

### Menu Interativo

Após conectar, você verá um menu com as seguintes opções:

```
=== Menu ===
1. Criar Database
2. Criar Usuário
3. Listar Databases
4. Sair
```

### Exemplo de Uso

1. **Criar uma Database**

   - Escolha opção 1
   - Digite o nome da database (ex: `myapp`)

2. **Criar um Usuário**

   - Escolha opção 2
   - Digite o nome da database (ex: `myapp`)
   - **Se a database não existir**, o programa perguntará se você deseja criá-la:
     - Digite `S` ou Enter para criar automaticamente
     - Digite `n` para cancelar a operação
   - Digite o nome do usuário (ex: `myuser`)
   - Digite a senha (pode conter caracteres especiais como @, #, %, etc.)
   - Digite as roles separadas por vírgula (ex: `read,readWrite`)
   - **O programa irá gerar automaticamente a connection string** com:
     - Senha com escape correto de caracteres especiais
     - Host e porta extraídos da URL original de conexão

3. **Listar Databases**
   - Escolha opção 3
   - Visualize todas as databases disponíveis

## Roles Disponíveis

- `read` - Leitura apenas
- `readWrite` - Leitura e escrita
- `dbAdmin` - Administração da database
- `userAdmin` - Gerenciamento de usuários
- `dbOwner` - Proprietário completo da database

## Exemplos de URLs de Conexão

Ao executar o programa, você pode usar qualquer um destes formatos:

### MongoDB Local

```
mongodb://localhost:27017
```

### MongoDB com Autenticação

```
mongodb://admin:password@localhost:27017
```

### MongoDB Atlas

```
mongodb+srv://user:password@cluster.mongodb.net/
```

### MongoDB com Opções

```
mongodb://localhost:27017/?authSource=admin
```

**Nota de Segurança**: A URL é solicitada dentro do programa para evitar que credenciais fiquem salvas no histórico do terminal.

## Exemplo de Saída ao Criar Usuário

### Quando a database já existe:

```
Nome da database: myapp
Nome do usuário: appuser
Senha: P@ss#w0rd!
Roles (separadas por vírgula, ex: read,readWrite): readWrite

✓ Usuário 'appuser' criado com sucesso na database 'myapp'!
   Roles atribuídas: [readWrite]

📋 Connection String:
mongodb://appuser:P%40ss%23w0rd%21@localhost:27017/myapp
```

### Quando a database não existe:

```
Nome da database: newdb
⚠️  A database 'newdb' não existe.
Deseja criar a database agora? (S/n): S
✓ Database 'newdb' criada com sucesso!
Nome do usuário: newuser
Senha: mypassword
Roles (separadas por vírgula, ex: read,readWrite):

✓ Usuário 'newuser' criado com sucesso na database 'newdb'!
   Roles atribuídas: [readWrite]

📋 Connection String:
mongodb://newuser:mypassword@localhost:27017/newdb
```

**Nota**: Caracteres especiais como `@`, `#`, `!`, `%` são automaticamente escapados na URL.

### Exemplos de Caracteres Especiais que são Escapados

| Caractere | Escapado como |
| --------- | ------------- |
| @         | %40           |
| #         | %23           |
| !         | %21           |
| $         | %24           |
| %         | %25           |
| &         | %26           |
| /         | %2F           |
| :         | %3A           |

## Desenvolvimento

### Requisitos

- Go 1.21 ou superior
- MongoDB 4.0 ou superior

### Executar sem Build

```bash
go run main.go
```

O programa solicitará a URL de conexão.

## Observações

- **Segurança**: A URL de conexão é solicitada dentro do programa, não como argumento de linha de comando, evitando que credenciais fiquem salvas no histórico do terminal
- Para criar usuários, você precisa estar conectado com um usuário que tenha permissões de `userAdmin` ou `userAdminAnyDatabase`
- No MongoDB, databases são criadas automaticamente ao inserir dados, mas este programa cria uma collection temporária para garantir que a database exista

## Licença

MIT
