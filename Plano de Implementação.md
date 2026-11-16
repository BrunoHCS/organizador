# Plano de Implementação – Biblioteca do SSD

*App portátil, estilo Netflix com identidade de educação, para organizar e consumir fotos, vídeos, músicas, PDFs, videoaulas, livros, ebooks e apostilas.*

## 1. Visão e objetivos

**Visão:** um app que mora dentro do SSD, abre em qualquer PC com Windows sem instalação e transforma uma pasta de arquivos em uma "plataforma de streaming de estudos" pessoal.

**Objetivos**

- Organizar automaticamente os arquivos do SSD (fotos, vídeos, músicas, PDFs, ebooks).
- Assistir videoaulas com progresso salvo, retomada de onde parou e próxima aula automática.
- Ler livros, ebooks e apostilas (EPUB e PDF) com marcadores, destaques e anotações.
- Ter uma interface agradável, rápida e familiar (carrosséis, banner de destaque, cards).
- Funcionar offline, sem conta, sem nuvem.

**Fora de escopo (por enquanto):** multiusuário, sincronização em nuvem, app mobile, DRM.

## 2. Requisitos

### Funcionais

| ID | Requisito | Prioridade |
| --- | --- | --- |
| RF01 | Varrer o SSD e catalogar arquivos (incremental) | Alta |
| RF02 | Navegar por pastas e por tipo | Alta |
| RF03 | Visualizar fotos em grade, com tela cheia e EXIF | Alta |
| RF04 | Reproduzir vídeos (velocidade, legendas, atalhos, PiP) | Alta |
| RF05 | Reproduzir músicas com fila e playlists | Média |
| RF06 | Ler PDF e EPUB com progresso salvo | Alta |
| RF07 | Agrupar vídeos/PDFs em **Cursos** (pasta = curso) com módulos e aulas | Alta |
| RF08 | "Continuar assistindo/lendo" | Alta |
| RF09 | Busca global (nome, metadados, tags, texto de PDFs) | Alta |
| RF10 | Favoritos, tags e coleções | Média |
| RF11 | Anotações com timestamp (vídeo) e destaques (livros) | Média |
| RF12 | Metas de estudo, histórico e estatísticas | Baixa |
| RF13 | Detecção de duplicados | Baixa |

### Não funcionais

- **Portabilidade:** um `.exe` + pasta de dados, sem instalador e sem CGO.
- **Desempenho:** abrir a home em menos de 1 s com 100 mil itens indexados; rolagem a 60 fps.
- **Resiliência:** sobreviver à remoção do SSD sem ejetar (catálogo íntegro).
- **Privacidade:** servidor apenas em `127.0.0.1`, sem telemetria.
- **Compatibilidade:** Windows 10/11, navegadores baseados em Chromium (Edge/Chrome) como alvo principal.

## 3. Design de UI/UX (estilo Netflix, identidade educacional)

### 3.1 Conceito

Layout e interações da Netflix (banner, carrosséis horizontais, cards que crescem no hover, tela de detalhes, player imersivo), mas com visual **calmo, focado e acolhedor**, que transmite estudo e confiança em vez do vermelho de entretenimento.

### 3.2 Paleta de cores

| Papel | Cor | Uso |
| --- | --- | --- |
| Fundo | `#0B1220` (azul-marinho profundo) | Fundo geral (tema escuro padrão) |
| Superfície | `#121C2F` / `#1A2740` | Cards, painéis, menus |
| Primária | `#3B82F6` (azul) | Botões principais, links, foco |
| Secundária | `#14B8A6` (teal) | Progresso, concluído, badges de curso |
| Destaque | `#F59E0B` (âmbar) | Favoritos, "continuar", marcadores e destaques de leitura |
| Texto | `#E8EEF9` / `#9FB0CC` | Principal / secundário |
| Erro | `#EF4444` | Alertas |

Tema claro opcional (fundo `#F6F8FC`, mesma primária) e **modo sépia** no leitor de livros.

### 3.3 Tipografia e ícones

- Interface: **Inter** (ou Poppins nos títulos de seção).
- Leitor de livros: **Literata** ou **Merriweather**, com tamanho, entrelinha e margens ajustáveis.
- Ícones: Lucide. Cantos arredondados (12 px nos cards) e sombras suaves.
- Fontes embutidas no binário (nada de CDN, o app é offline).

### 3.4 Telas

1. **Home**: banner de destaque (curso em andamento ou aleatório), carrosséis:
   - Continuar de onde parou
   - Meus cursos
   - Adicionados recentemente
   - Livros e apostilas
   - Favoritos
   - Músicas e playlists
   - Fotos recentes
2. **Cursos**: grade de cursos com capa, barra de progresso e nº de aulas.
3. **Detalhe do curso**: capa grande, descrição, botão "Continuar", lista de módulos e aulas com duração, check de concluída e materiais anexos (PDFs da pasta).
4. **Player de vídeo**: tela cheia imersiva, lista lateral de aulas, painel de anotações, velocidade (0,5x a 2x), legendas, atalhos de teclado, retomada automática, autoplay da próxima aula.
5. **Biblioteca de livros**: estantes com capas, filtro por autor/tag/status (não lido, lendo, lido).
6. **Leitor**: EPUB e PDF, índice lateral, busca no texto, marcadores, destaques coloridos, anotações, tema claro/escuro/sépia, progresso em %.
7. **Fotos**: grade por data (linha do tempo), visualizador com zoom, navegação por setas, EXIF.
8. **Músicas**: lista por álbum/artista, mini-player persistente no rodapé.
9. **Explorar pastas**: árvore do SSD (modo "arquivos"), para o caso de querer ver tudo cru.
10. **Busca**: resultados agrupados por tipo, com filtros.
11. **Estatísticas/Metas** (fase final): horas estudadas, streak, páginas lidas.
12. **Configurações**: tema, pastas ignoradas, reindexação, atalhos, backup do catálogo.

### 3.5 Interações-chave

- Hover no card: expande com título, duração/páginas, barra de progresso e botões (play, favoritar, detalhes).
- Navegação por teclado completa (setas, espaço, `F` tela cheia, `J/K/L` no vídeo, `Esc`).
- Skeletons de carregamento e miniaturas lazy-load.
- Responsivo (de 1280 px até telas grandes; tablet como bônus).

## 4. Arquitetura

```
┌────────────────────── Navegador (Edge/Chrome) ──────────────────────┐
│  SPA React + TypeScript + Tailwind  (embutida no binário)           │
└───────────────▲─────────────────────────────────────────────────────┘
                │ HTTP (127.0.0.1:porta) + Range requests
┌───────────────┴─────────────── app.exe (Go) ───────────────────────┐
│  API REST  │  Streaming de mídia  │  Scanner/Indexador  │  Jobs    │
│  SQLite (modernc, sem CGO) + FTS5   │  Cache de miniaturas         │
│  Processos auxiliares: ffmpeg / ffprobe (empacotados)               │
└─────────────────────────────────────────────────────────────────────┘
                │
        SSD:\Organizador\data\catalogo.db   SSD:\Organizador\data\thumbs\
```

### 4.1 Stack

| Camada | Escolha | Motivo |
| --- | --- | --- |
| Backend | Go 1.22+ (`net/http` + `chi`) | Binário único, rápido, fácil cross-compile |
| Banco | SQLite via `modernc.org/sqlite` | Sem CGO, portátil, com FTS5 |
| Frontend | React + Vite + TypeScript + Tailwind | Produtividade e ecossistema |
| Estado/dados | TanStack Query + Zustand | Cache de API e estado de UI |
| Vídeo | `<video>` HTML5 (+ `hls.js` se houver transcodificação) | Nativo e leve |
| PDF | PDF.js | Padrão de mercado |
| EPUB | epub.js (ou foliate-js) | Renderização e CFI para progresso |
| Mídia | ffmpeg/ffprobe | Miniaturas, metadados, conversão |
| Metadados de imagem/áudio | `rwcarlsen/goexif`, `dhowden/tag` | EXIF e ID3 |
| Empacotamento | `go:embed` + GoReleaser | Frontend dentro do `.exe` |

### 4.2 Decisões importantes

- **Caminhos relativos à raiz do app**: o `.exe` descobre onde está e trata a pasta-pai como raiz; a letra do drive pode mudar.
- **SQLite com `journal_mode=DELETE`** (não WAL) e `synchronous=FULL`: mais seguro em disco removível. Fechar o banco no encerramento.
- **Backup automático** do `catalogo.db` a cada N execuções (rotação de 3 cópias).
- **Streaming com HTTP Range** para vídeos e áudios (seek instantâneo).
- **Segurança local:** bind em `127.0.0.1`, porta aleatória, token de sessão no primeiro acesso, validação anti *path traversal* em todo acesso a arquivo.
- **Instância única:** lock file para impedir duas execuções no mesmo SSD.

## 5. Modelo de dados (SQLite)

```sql
CREATE TABLE items (
  id INTEGER PRIMARY KEY,
  rel_path TEXT UNIQUE NOT NULL,   -- relativo à raiz
  name TEXT NOT NULL,
  ext TEXT, type TEXT NOT NULL,    -- photo | video | audio | pdf | ebook | other
  size INTEGER, mtime INTEGER,
  hash TEXT,                       -- parcial (início+meio+fim) p/ duplicados
  duration_s REAL, width INTEGER, height INTEGER, pages INTEGER,
  meta_json TEXT,                  -- EXIF, ID3, autor, título etc.
  thumb_path TEXT,
  added_at INTEGER, deleted INTEGER DEFAULT 0
);

CREATE TABLE collections (          -- cursos, séries, álbuns, estantes
  id INTEGER PRIMARY KEY,
  kind TEXT NOT NULL,              -- course | album | shelf | manual
  title TEXT, description TEXT, cover_item_id INTEGER,
  root_rel_path TEXT, auto INTEGER DEFAULT 1
);

CREATE TABLE collection_items (
  collection_id INTEGER, item_id INTEGER,
  module TEXT, position INTEGER,   -- módulo e ordem da aula
  PRIMARY KEY (collection_id, item_id)
);

CREATE TABLE progress (
  item_id INTEGER PRIMARY KEY,
  position REAL,                   -- segundos (vídeo/áudio) ou página
  locator TEXT,                    -- CFI do EPUB
  percent REAL, completed INTEGER DEFAULT 0, updated_at INTEGER
);

CREATE TABLE notes (
  id INTEGER PRIMARY KEY, item_id INTEGER,
  kind TEXT,                       -- note | highlight | bookmark
  position REAL, locator TEXT, text TEXT, color TEXT, created_at INTEGER
);

CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT UNIQUE);
CREATE TABLE item_tags (item_id INTEGER, tag_id INTEGER, PRIMARY KEY(item_id, tag_id));
CREATE TABLE favorites (item_id INTEGER PRIMARY KEY, created_at INTEGER);
CREATE TABLE study_log (day TEXT, item_id INTEGER, seconds INTEGER, pages INTEGER);
CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT);

CREATE VIRTUAL TABLE search USING fts5(name, title, author, content, item_id UNINDEXED);
```

## 6. Módulos do backend

### 6.1 Scanner / Indexador

- Percorre a raiz com `filepath.WalkDir`, ignorando pastas configuradas (`$RECYCLE.BIN`, `System Volume Information`, a própria pasta `Organizador`).
- **Incremental:** compara `size + mtime`; só reprocessa o que mudou; marca como `deleted` o que sumiu.
- **Pipeline com workers** (ex.: 4 goroutines): classificar → extrair metadados → gerar miniatura → indexar texto.
- **Progresso em tempo real** para a UI via Server-Sent Events.
- Opcional: `fsnotify` para detectar mudanças enquanto o app está aberto.

### 6.2 Detecção de cursos (regra automática)

1. Pasta que contém vídeos (ou vídeos + PDFs) vira **curso**.
2. Subpastas viram **módulos**; arquivos viram **aulas**.
3. Ordenação **natural** (`1`, `2`, `10` e não `1`, `10`, `2`), reconhecendo prefixos como `01 -`, `Aula 03`, `Módulo 2`.
4. Capa: `cover.jpg`/`folder.jpg` se existir; senão, frame do primeiro vídeo.
5. PDFs e anexos na pasta viram **materiais de apoio** da aula/curso.
6. Legendas `.srt`/`.vtt` com o mesmo nome do vídeo são associadas e convertidas para VTT.
7. Arquivo opcional `curso.json` na pasta permite sobrescrever título, descrição, ordem e capa.

### 6.3 Mídia e miniaturas

- **Fotos:** miniaturas em JPEG/WebP (320 px e 1080 px); HEIC convertido via ffmpeg.
- **Vídeos:** `ffprobe` (duração, codec, capítulos, faixas de áudio/legenda); miniatura de um frame a \~10% e *sprite sheet* opcional para preview na barra de progresso.
- **Transcodificação sob demanda** (HEVC/MKV que o navegador não toca): ffmpeg gera H.264/AAC (HLS ou MP4 fragmentado) e guarda em cache.
- **Áudio:** tags ID3, capa embutida.
- **PDF:** capa da primeira página (via PDF.js no cliente ou `pdftoppm`), contagem de páginas.
- **EPUB:** capa, título e autor lidos do OPF (é um ZIP, simples de ler em Go).

### 6.4 Progresso e estudo

- A UI envia posição a cada \~5 s (debounce) e ao fechar/pausar.
- Vídeo concluído quando passa de 90% (configurável).
- `study_log` acumula tempo e páginas por dia, base para metas e streak.

### 6.5 Busca

- FTS5 sobre nome, título, autor e **texto extraído de PDFs/EPUBs** (extração em segundo plano, com limite de tamanho).
- Ranking por relevância com *boost* para título; destaque do trecho (`snippet()`).

## 7. API (resumo)

| Método | Rota | Função |
| --- | --- | --- |
| GET | `/api/home` | Carrosséis da home |
| GET | `/api/items?type=&q=&tag=&page=` | Listagem paginada |
| GET | `/api/items/{id}` | Detalhe e metadados |
| GET | `/media/{id}` | Stream do arquivo (Range) |
| GET | `/media/{id}/transcoded` | Stream transcodificado |
| GET | `/thumb/{id}?size=` | Miniatura |
| GET | `/api/courses`, `/api/courses/{id}` | Cursos, módulos e aulas |
| GET/PUT | `/api/progress/{itemId}` | Ler/salvar progresso |
| GET/POST/DELETE | `/api/notes` | Anotações, destaques, marcadores |
| POST/DELETE | `/api/favorites/{id}` | Favoritos |
| GET/POST | `/api/tags` | Tags |
| GET | `/api/search?q=` | Busca global |
| POST | `/api/scan` | Inicia varredura |
| GET | `/api/scan/events` | Progresso (SSE) |
| GET | `/api/stats` | Estatísticas de estudo |
| GET | `/healthz` | Health check |

## 8. Estrutura do repositório

```
biblioteca-ssd/
├── cmd/app/main.go            # bootstrap, flags, abre o navegador
├── internal/
│   ├── config/                # raiz, caminhos, settings
│   ├── db/                    # conexão, migrations, queries
│   ├── scanner/               # walk, classificação, hash
│   ├── courses/               # regras de detecção de cursos
│   ├── media/                 # ffprobe, thumbs, transcode
│   ├── reader/                # epub/pdf metadados e texto
│   ├── search/                # FTS5
│   ├── api/                   # handlers, middlewares, SSE
│   └── security/              # token, path guard
├── web/                       # React + Vite
│   └── src/{pages,components,player,reader,hooks,styles}
├── assets/ffmpeg/             # ffmpeg.exe, ffprobe.exe (não versionar binário)
├── scripts/                   # build, release
├── .goreleaser.yaml
└── README.md
```

## 9. Roadmap por fases

Estimativas para desenvolvimento solo em tempo parcial (cerca de 8–10 h/semana).

### Fase 0 – Fundação (1 semana)

- Repositório, CI (build + testes), estrutura Go + Vite, `go:embed` do frontend.
- Migrations do SQLite, config, logs estruturados (`slog`), `/healthz`.
- Design system: tokens de cor/tipografia, componentes base (Card, Carousel, Button, Modal).
- **Pronto quando:** o `.exe` abre o navegador com uma página "olá" e cria o banco no SSD.

### Fase 1 – Catálogo e navegação (2 semanas)

- Scanner incremental, classificação por tipo, hash parcial, SSE de progresso.
- Tela "Explorar pastas" e listagem por tipo, com paginação.
- **Pronto quando:** indexar 50 mil arquivos em minutos e reabrir sem reescanear tudo.

### Fase 2 – Fotos e PDFs (2 semanas)

- Miniaturas (fotos, capa de PDF), grade em linha do tempo, visualizador com zoom e EXIF.
- Leitor de PDF (PDF.js) com progresso, marcadores e busca no documento.
- **Pronto quando:** abrir uma foto/PDF e voltar à última página ao reabrir.

### Fase 3 – Vídeo e cursos (2–3 semanas)  ← núcleo do valor

- Streaming com Range, player customizado (velocidade, legendas, atalhos, PiP, retomada).
- Detecção automática de cursos, módulos e aulas; tela de curso; autoplay da próxima aula.
- Home estilo Netflix com "Continuar assistindo" e "Meus cursos".
- Transcodificação sob demanda para formatos incompatíveis.
- **Pronto quando:** um curso real do SSD aparece como curso, toca do início ao fim e lembra o progresso.

### Fase 4 – Livros e ebooks (2 semanas)

- Leitor de EPUB (índice, temas claro/escuro/sépia, fonte ajustável, CFI de progresso).
- Estante de livros com capas, status de leitura, destaques e anotações.
- **Pronto quando:** ler um EPUB e um PDF com destaques persistidos e progresso em %.

### Fase 5 – Músicas, busca e organização (2 semanas)

- Player de áudio com fila, mini-player persistente, playlists.
- Busca global com FTS5 (incluindo texto de PDFs/EPUBs), tags, favoritos e coleções manuais.
- **Pronto quando:** achar um termo dentro de um PDF a partir da busca global.

### Fase 6 – Estudo e polimento (2 semanas)

- Anotações com timestamp no vídeo, metas de estudo, estatísticas e streak.
- Detecção de duplicados, backup do catálogo, tela de configurações.
- Acessibilidade (contraste, foco, `aria`), performance (virtualização de listas), tratamento de erros.
- **Pronto quando:** usar o app por uma semana inteira sem travar ou perder progresso.

### Fase 7 – Empacotamento e release (1 semana)

- Build de release com GoReleaser, pasta final pronta para copiar ao SSD, ícone do app, script de atalho.
- Documentação de uso e checklist de restauração do catálogo.

**Total estimado:** 12 a 14 semanas em tempo parcial. Um MVP utilizável (Fases 0 a 3) fica pronto em cerca de 6 a 8 semanas.

## 10. Qualidade e testes

- **Unitários (Go):** ordenação natural, detecção de cursos, classificação de tipos, anti *path traversal*.
- **Integração:** scanner em uma pasta-fixture com arquivos de amostra (golden files do resultado esperado).
- **E2E:** Playwright cobrindo home → curso → player → retomada de progresso.
- **Desempenho:** benchmark do scanner (arquivos/s) e da listagem (latência da API com 100 mil itens).
- **Resiliência:** teste de "puxar o SSD" (matar o processo durante a escrita) e validar `PRAGMA integrity_check` ao abrir.

## 11. Observabilidade (do jeito SRE)

- Logs estruturados em `data/logs/app.log` com rotação.
- Métricas simples no log: duração do scan, arquivos/s, fila de miniaturas, erros de ffmpeg.
- `/healthz` e `/debug/stats` (somente local) para diagnóstico.
- Tela "Diagnóstico" nas configurações: versão, tamanho do banco, itens por tipo, últimos erros.

## 12. Empacotamento e distribuição (Windows)

- `GOOS=windows GOARCH=amd64 go build` com `-ldflags "-s -w"`; frontend embutido via `go:embed`.
- Pasta final: `Organizador\app.exe`, `ffmpeg\`, `data\` (criada no primeiro uso) e um atalho `Abrir Biblioteca.lnk`.
- Binário sem assinatura dispara SmartScreen na primeira execução (aceitável para uso pessoal).
- Atualização: substituir o `app.exe`; as migrations rodam sozinhas ao abrir.

## 13. Riscos e mitigações

| Risco | Impacto | Mitigação |
| --- | --- | --- |
| Corrupção do banco ao remover o SSD | Alto | `journal_mode=DELETE`, backups rotativos, `integrity_check` na abertura, reconstrução a partir do scan |
| Vídeos HEVC/MKV não tocam no navegador | Alto | Transcodificação sob demanda com cache |
| Varredura lenta em muitos arquivos | Médio | Incremental, workers paralelos, hash parcial |
| Nomes fora de padrão quebram a ordem das aulas | Médio | Ordenação natural + `curso.json` para ajuste manual |
| Escopo grande demais | Alto | Entregar por fases; MVP = Fases 0–3 |
| Fonte de texto de PDFs escaneados (sem texto) | Baixo | OCR opcional no futuro (Tesseract) |

## 14. Backlog futuro

- OCR de PDFs escaneados e busca dentro de imagens.
- Reconhecimento de rostos e álbuns inteligentes nas fotos.
- Flashcards a partir de destaques e anotações (repetição espaçada).
- Exportar anotações para Markdown/Obsidian.
- Modo "app" com Tauri/WebView2 e ícone na bandeja.
- Acesso pela rede local (com senha) para assistir no celular ou na TV.
- Legendas automáticas com Whisper local.

## 15. Próximos passos imediatos

1. Confirmar o sistema de arquivos do SSD (NTFS ou exFAT) e o espaço livre.
2. Separar uma pasta de amostra (1 curso, alguns livros, fotos) para testar desde a Fase 1.
3. Criar o repositório e fechar a Fase 0.
4. Prototipar a Home no Figma (ou direto em código) com a paleta da seção 3.
