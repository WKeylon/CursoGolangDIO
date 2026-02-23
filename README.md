# Sistema de Pesquisa Eleitoral

Este é um sistema de pesquisa eleitoral desenvolvido em Go usando a biblioteca Fyne para a interface gráfica.

## Pré-requisitos

Para compilar e rodar a aplicação, você precisa ter instalado:

- Go 1.16 ou superior
- Bibliotecas de desenvolvimento gráfico (necessárias para o Fyne):
    - **Linux (Debian/Ubuntu):** `sudo apt-get install libgl1-mesa-dev xorg-dev`
    - **Windows/macOS:** Geralmente não é necessário instalar dependências extras além do compilador C (como GCC ou Mingw).

## Como Executar

1. Navegue até o diretório do projeto.
2. Execute o comando:

```bash
go run ./cmd/app
```

## Funcionalidades

- **Votação**: Permite selecionar um candidato e responder perguntas de Sim/Não.
- **Estatísticas**: Mostra o total de votos por candidato e as respostas das perguntas.
- **Administração**: Permite adicionar novos candidatos e perguntas.
