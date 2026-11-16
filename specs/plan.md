# Plano de Implementação — Biblioteca do SSD

## Direção técnica

Construir um aplicativo local portátil: backend Go com API HTTP e streaming, banco SQLite sem CGO e frontend SPA React embutido no executável. O diretório de dados acompanha o executável no SSD. A implementação segue as fases abaixo para entregar valor cedo e manter o MVP nas Fases 0–3.

### Estrutura de destino

```text
biblioteca-ssd/
├── cmd/app/main.go
├── internal/
│   ├── config/       # raiz, caminhos, configurações
│   ├── db/           # conexão, migrations, consultas
│   ├── scanner/      # walk, classificação, hash
│   ├── courses/      # detecção de cursos
│   ├── media/        # ffprobe, miniaturas, transcodificação
│   ├── reader/       # metadados e texto de PDF/EPUB
│   ├── search/       # FTS5
│   ├── api/          # handlers, middleware e SSE
│   └── security/     # token e proteção de caminhos
├── web/              # React/Vite; web/src/pages, components, player, reader, hooks, styles
├── assets/ffmpeg/    # ffmpeg.exe e ffprobe.exe, não versionados
├── scripts/
├── specs/
└── README.md
```

O diretório `src/` solicitado na raiz será o espaço inicial para o código; ao iniciar a implementação, sua organização deve convergir para os caminhos `cmd/`, `internal/` e `web/` acima, preservando `web/src/` para a aplicação frontend.

## Fases e marcos

| Fase | Estimativa | Entrega e critério de conclusão |
| --- | --- | --- |
| 0 — Fundação | 1 semana | Repositório, CI de build/testes, esqueleto Go + Vite, frontend embutido, migrations SQLite, configuração, logs `slog`, `/healthz` e tokens visuais/componentes base. **Pronto quando:** o `.exe` abre o navegador na página inicial e cria o banco no SSD. |
| 1 — Catálogo e navegação | 2 semanas | Scanner incremental, classificação, hash parcial, SSE, explorar pastas e listagem paginada por tipo. **Pronto quando:** indexa 50 mil arquivos em minutos e reabre sem revarrer tudo. |
| 2 — Fotos e PDFs | 2 semanas | Miniaturas, capa PDF, grade por data, visualizador de foto com zoom/EXIF e leitor PDF.js com progresso, marcadores e busca. **Pronto quando:** foto/PDF abrem e o PDF retorna à última página. |
| 3 — Vídeo e cursos | 2–3 semanas | Streaming Range, player, retomada, legendas, atalhos, cursos/módulos/aulas, detalhes, autoplay, home e transcodificação sob demanda. **Pronto quando:** curso real toca do início ao fim e progresso é lembrado. |
| 4 — Livros e ebooks | 2 semanas | Leitor EPUB, CFI, índice, temas e fontes ajustáveis; estante, destaques e notas. **Pronto quando:** EPUB e PDF persistem destaques e progresso percentual. |
| 5 — Música, busca e organização | 2 semanas | Player/fila/mini-player, FTS5 em texto PDF/EPUB, tags, favoritos e coleções manuais. **Pronto quando:** busca global localiza termo dentro de PDF. |
| 6 — Estudo e polimento | 2 semanas | Notas temporizadas, metas, estatísticas, streak, duplicados, backup, configurações, acessibilidade, virtualização e tratamento de falhas. **Pronto quando:** uso contínuo por uma semana sem travamento ou perda de progresso. |
| 7 — Empacotamento e release | 1 semana | GoReleaser, executável e recursos para SSD, ícone, atalho, documentação e restauração. **Pronto quando:** pasta de release pode ser copiada e aberta em Windows compatível. |

**Estimativa total:** 12–14 semanas em tempo parcial (8–10 h/semana). MVP das Fases 0–3: aproximadamente 6–8 semanas.

## Decisões de arquitetura a manter

- Dados em `data/` ao lado do executável e caminhos relativos à pasta do app; a raiz da biblioteca é escolhida pelo usuário e a letra da unidade não faz parte da identidade do catálogo.
- SQLite com `journal_mode=DELETE`, `synchronous=FULL`, fechamento limpo, `integrity_check` na abertura e três backups rotativos.
- HTTP Range para seek em vídeo/áudio; SSE para progresso de scan.
- Token local, porta aleatória, bind estrito somente em `127.0.0.1`, validação de caminhos e lock de instância única.
- Frontend sem CDN; binários ffmpeg/ffprobe junto do app e fora do controle de versão.

## Qualidade e validação

- Unitários planejados: ordenação natural, detecção de curso, classificação e proteção contra *path traversal*.
- Integração planejada: scanner em pasta fixture e resultado esperado.
- E2E planejado: home → curso → player → retomada.
- Desempenho planejado: taxa de scan e latência de listagem com 100 mil itens.
- Resiliência planejada: interrupção durante escrita seguida de `PRAGMA integrity_check`.
- Observabilidade com logs estruturados, métricas simples e diagnóstico local.

## Riscos e respostas

| Risco | Resposta planejada |
| --- | --- |
| Remoção do SSD durante escrita | SQLite em modo seguro, backups rotativos, checagem na abertura e reindexação reconstruível |
| HEVC/MKV não reproduzido pelo navegador | Transcodificação sob demanda com cache |
| Scan lento | Incrementalidade, workers paralelos e hash parcial |
| Nome irregular altera ordem de aula | Ordenação natural e metadados opcionais em `curso.json` |
| Escopo excessivo | Priorizar Fases 0–3 como MVP e adiar backlog |
| PDF escaneado sem texto | OCR opcional futuro via Tesseract |

## Backlog após a versão inicial

OCR e busca em imagens; reconhecimento facial e álbuns inteligentes; flashcards com repetição espaçada; exportação de notas para Markdown/Obsidian; modo desktop com Tauri/WebView2; legendas automáticas com Whisper local. Acesso pela rede local permanece fora do escopo; o servidor escuta somente em `127.0.0.1`.
