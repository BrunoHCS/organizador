# Organizador

Aplicativo local e portátil para transformar um SSD em uma biblioteca pessoal de estudos, com navegação, reprodução e organização de arquivos multimídia, documentos e cursos — tudo funcionando offline, sem conta e sem dependência de nuvem.

## Visão geral

O projeto foi concebido para pessoas que guardam materiais de estudo em um SSD e desejam acessar tudo de forma simples, rápida e organizada em computadores Windows 10/11. A aplicação funciona como uma biblioteca digital local: cataloga arquivos automaticamente, organiza cursos e materiais, preserva progresso, permite leitura de PDFs/EPUBs, reprodução de vídeos, fotos e música, além de manter buscas e anotações em um ambiente privativo.

Este repositório centraliza a documentação e o planejamento de desenvolvimento do produto, que segue uma arquitetura de backend em Go, frontend em React + Vite e banco SQLite local.

## Status do projeto

Em desenvolvimento.

O escopo atual está alinhado com a especificação técnica do produto e com a estratégia da implementação em fases, priorizando um MVP funcional nas fases iniciais.

## Objetivos principais

- Catalogar arquivos do SSD de forma incremental e segura
- Navegar por pastas, tipos de mídia e cursos
- Reproduzir vídeos, fotos, música, PDFs e EPUBs no ambiente local
- Salvar progresso de leitura e reprodução
- Buscar por nome, metadados e texto extraído de documentos
- Manter a organização, anotações, favoritos e tags sem depender de internet
- Funcionar como um aplicativo portátil, sem instalador e sem conta de usuário

## Funcionalidades previstas

### Biblioteca e catálogo
- Varredura automática da raiz escolhida pelo usuário
- Organização por tipo: foto, vídeo, áudio, PDF, EPUB, cursos e outros
- Catalogação incremental com atualização de arquivos alterados
- Suporte a pastas, coleções e cursos
- Deteção de duplicados por hash parcial

### Visualização e leitura
- Grade de fotos com zoom e navegação
- Reprodução de vídeos com seek, legendas, tela cheia e retomada
- Player de música com fila e mini-player persistente
- Leitura de PDFs e EPUBs com progresso, marcadores, destaques e anotações
- Página de “Continuar assistindo/lendo” com retomada automática

### Organização e estudo
- Favoritos, tags e coleções manuais
- Notas com marcação de tempo em vídeos
- Histórico de estudo e estatísticas
- Busca global por nome, metadados e conteúdo de documentos
- Ajuste de configurações do ambiente e da biblioteca

## Arquitetura proposta

### Backend
- Go 1.22+
- API HTTP em `net/http`
- SQLite com FTS5 e persistência local
- Scanner/catálogo e processamento de mídia
- Streaming de arquivos com suporte a `Range`
- Segurança local com bind em `127.0.0.1`, token de sessão e validação de caminhos

### Frontend
- React
- Vite
- TypeScript
- Tailwind CSS
- Zustand e TanStack Query para estado e dados da interface

### Persistência e dados
- Banco em `data/catalogo.db`
- Miniaturas em `data/thumbs/`
- Arquivos do usuário armazenados junto ao executável
- Backups e checagem de integridade para resiliência

## Estrutura de diretórios planejada

```text
organizador/
├── cmd/
│   └── app/
├── internal/
│   ├── api/
│   ├── config/
│   ├── courses/
│   ├── db/
│   ├── media/
│   ├── reader/
│   ├── scanner/
│   ├── search/
│   └── security/
├── web/
│   └── src/
├── assets/
├── scripts/
├── specs/
├── LICENSE
├── Plano de Implementação.md
├── README.md
└── src/
```

## Requisitos

### Sistema operacional
- Windows 10 ou 11

### Ambiente de desenvolvimento
- Go 1.22+
- Node.js 18+
- npm ou pnpm
- ffmpeg e ffprobe (para transcodificação e metadados, quando aplicável)

## Como executar

A execução ainda será definida conforme a implementação evoluir. O objetivo final é um executável portátil para Windows, com dados do aplicativo armazenados ao lado do binário e a biblioteca do usuário escolhida pela pessoa que usa o sistema.

### Fluxo esperado
1. Baixar ou compilar o executável
2. Abrir a aplicação no Windows
3. Escolher a raiz da biblioteca a catalogar
4. Iniciar a varredura inicial
5. Navegar por fotos, vídeos, músicas, PDFs, EPUBs e cursos
6. Continuar estudos e progresso já salvos automaticamente

## Roadmap

O projeto segue a seguinte sequência de desenvolvimento:

- Fase 0: base da aplicação, configuração, banco e estrutura inicial
- Fase 1: catálogo, varredura e navegação
- Fase 2: fotos e PDFs
- Fase 3: vídeo e cursos
- Fase 4: livros e EPUBs
- Fase 5: música, busca e organização
- Fase 6: estudo, estatísticas e refinamento
- Fase 7: empacotamento e release

## Documentação de especificação

A base funcional e técnica do produto está documentada nos arquivos em `specs/`:

- `specs/spec.md` — especificação funcional e regras de domínio
- `specs/plan.md` — plano de implementação e arquitetura
- `specs/tasks.md` — organização das tarefas do projeto
- `specs/constitution.md` — princípios e diretrizes do produto

## Licença

Este projeto está sob a licença indicada no arquivo `LICENSE`.

## Observações

- O sistema foi pensado para operar offline e preservar privacidade
- O uso é local apenas em `127.0.0.1`
- O objetivo principal é transformar um SSD em uma biblioteca de estudo portátil, confiável e fácil de manter
