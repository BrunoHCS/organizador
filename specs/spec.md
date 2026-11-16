# Especificação — Biblioteca do SSD

## 1. Resumo

Aplicativo portátil para Windows que transforma os arquivos de um SSD em uma biblioteca pessoal de estudos, com navegação e consumo de fotos, vídeos, músicas, PDFs, EPUBs, livros e apostilas. Deve funcionar offline, sem conta, sem nuvem e sem instalador.

## 2. Usuários e contexto

- Pessoa que conecta o SSD a computadores Windows 10/11 e quer organizar e estudar a partir dos próprios arquivos.
- Uso principalmente em Edge ou Chrome, por meio de uma SPA local aberta pelo executável.
- A letra da unidade pode mudar entre computadores; o aplicativo e o catálogo devem continuar encontráveis.
- No primeiro uso, a pessoa escolhe a raiz da biblioteca a catalogar; os dados do aplicativo ficam na pasta `data/` ao lado do executável.

## 3. Objetivos

1. Catalogar arquivos automaticamente com varredura incremental.
2. Navegar por pastas, tipos, cursos, livros, fotos e músicas.
3. Retomar vídeos e leituras de onde pararam e acompanhar aulas.
4. Encontrar arquivos por nome, metadados, tags e texto de PDFs/EPUBs.
5. Preservar catálogo, progresso e anotações no SSD com segurança.

## 4. Fora de escopo nesta versão

Multiusuário, sincronização em nuvem, aplicativo mobile, DRM, OCR obrigatório de PDFs escaneados, reconhecimento facial, flashcards, exportação para Obsidian, uso pela rede local e transcrição automática.

## 5. Requisitos funcionais

| ID | Requisito | Prioridade |
| --- | --- | --- |
| RF01 | Varrer a raiz da biblioteca escolhida e catalogar arquivos incrementalmente, atualizando itens alterados; marcar ausentes como removidos somente após uma varredura completa e sem erros. | Alta |
| RF02 | Navegar pela árvore de pastas e filtrar/listar itens por tipo, com paginação. | Alta |
| RF03 | Exibir fotos em grade/linha do tempo, abrir em tela cheia, permitir zoom e navegação por setas, e mostrar EXIF disponível. | Alta |
| RF04 | Reproduzir vídeos com seek, velocidade de 0,5x a 2x, legendas, atalhos, tela cheia, PiP quando suportado, retomada e próxima aula automática; garantir reprodução de HEVC/MKV, convertendo quando necessário. | Alta |
| RF05 | Reproduzir músicas com fila, playlists, agrupamento por álbum/artista e mini-player persistente. | Média |
| RF06 | Ler PDF e EPUB, salvar progresso, oferecer índice, busca, marcadores, destaques e anotações. | Alta |
| RF07 | Detectar cursos automaticamente segundo critérios explícitos, organizando subpastas como módulos e arquivos como aulas; permitir que o usuário corrija a classificação. | Alta |
| RF08 | Apresentar uma seção “Continuar assistindo/lendo” com retomada do item e posição salvos. | Alta |
| RF09 | Busca global por nome, metadados, tags e texto extraído de PDFs/EPUBs, com resultados agrupados por tipo. | Alta |
| RF10 | Permitir favoritos, tags e coleções manuais. | Média |
| RF11 | Criar notas com timestamp em vídeos e notas, destaques coloridos e marcadores em livros. | Média |
| RF12 | Registrar histórico, tempo/páginas estudados, metas e estatísticas; permitir ajustar o limiar de conclusão de vídeos (padrão de 90%) e reverter a conclusão. | Baixa |
| RF13 | Detectar possíveis duplicados por hash parcial. | Baixa |
| RF14 | Exibir configurações para tema, pastas ignoradas, reindexação, atalhos e backup do catálogo. | Média |

## 6. Regras de domínio

### 6.1 Varredura e catálogo

- Percorrer a raiz da biblioteca escolhida pelo usuário com `filepath.WalkDir` e ignorar `$RECYCLE.BIN`, `System Volume Information` e a pasta do próprio aplicativo por padrão.
- Comparar tamanho e data de modificação para reprocessar apenas o necessário; marcar itens que sumiram como removidos somente após uma varredura completa e sem erros.
- Pipeline com workers para classificar, extrair metadados, gerar miniaturas e indexar texto.
- Mostrar o progresso da varredura à interface via Server-Sent Events; `fsnotify` é opcional.
- Categorias iniciais: `photo`, `video`, `audio`, `pdf`, `ebook`, `other`.

### 6.2 Cursos

- Reconhecer uma pasta como curso somente quando cumprir critérios explícitos de detecção; esses critérios ainda precisam ser definidos. Permitir que o usuário corrija a classificação automática.
- Subpastas representam módulos; vídeos representam aulas; PDFs e anexos são materiais de apoio.
- Ordenação natural (`1`, `2`, `10`) reconhece prefixos como `01 -`, `Aula 03` e `Módulo 2`.
- `cover.jpg`/`folder.jpg` define a capa; na ausência, usar um frame do primeiro vídeo quando disponível.
- Legendas `.srt`/`.vtt` de mesmo nome são associadas; SRT pode ser convertido para VTT.
- `curso.json` opcional pode substituir título, descrição, ordem e capa.

### 6.3 Progresso e busca

- Salvar posição de reprodução/leitura aproximadamente a cada 5 segundos e ao pausar/fechar.
- O limiar de conclusão de vídeo é configurável, começa em 90% e pode ser revertido pelo usuário.
- Registrar tempo e páginas por dia para estatísticas e metas.
- FTS5 indexa nome, título, autor e texto extraído de PDF/EPUB em segundo plano, até um limite fixo definido pelo produto (valor a definir); indicar ao usuário os arquivos cujo conteúdo não foi indexado por exceder esse limite. Título recebe boost e resultados podem exibir snippet.

## 7. Requisitos não funcionais e critérios de aceite

- **Portabilidade:** executável Windows sem instalador, frontend embutido, sem CGO; a raiz da biblioteca é escolhida pelo usuário e os dados do aplicativo ficam em `data/`, ao lado do executável.
- **Offline e privacidade:** operação sem internet; servidor somente em `127.0.0.1`; sem telemetria.
- **Desempenho:** home abre em menos de 1 s com 100 mil itens; rolagem de listas a 60 fps; listagens paginadas/virtualizadas.
- **Resiliência:** banco íntegro após encerramento inesperado/remover SSD; backup rotativo de três cópias no mesmo SSD que contém os dados do aplicativo, incluindo catálogo, progresso, anotações, configurações e demais dados do usuário; `integrity_check` ao abrir.
- **Compatibilidade:** Windows 10/11; Edge/Chrome como navegadores alvo.
- **MVP:** ao final das Fases 0–3, o usuário consegue indexar o SSD, explorar pastas, visualizar fotos/PDFs, abrir cursos, assistir vídeo e retomar o progresso.

## 8. UX e interface

- Home com banner de destaque e carrosséis: Continuar, Meus cursos, Adicionados recentemente, Livros e apostilas, Favoritos, Músicas e playlists e Fotos recentes.
- Telas de cursos, detalhe de curso, player de vídeo, biblioteca/leitor de livros, fotos, músicas, explorar pastas, busca, estatísticas/metas e configurações.
- Cards expandem no hover com título, duração/páginas, progresso e ações relevantes.
- Navegação por teclado completa: setas, espaço, `F`, `J/K/L` no vídeo e `Esc`.
- Skeletons e miniaturas lazy-load; layout responsivo a partir de 1280 px, com tablet como bônus.
- Tema escuro padrão: fundo `#0B1220`, superfícies `#121C2F`/`#1A2740`, primária `#3B82F6`, secundária `#14B8A6`, destaque `#F59E0B`, texto `#E8EEF9`/`#9FB0CC`, erro `#EF4444`. Tema claro opcional e sépia no leitor.
- Interface em Inter (ou Poppins nos títulos); leitura em Literata (ou Merriweather), com fonte, entrelinha e margens ajustáveis; ícones Lucide, cards com raio de 12 px e sombras suaves.

## 9. Arquitetura e dados

- **Backend:** Go 1.22+, `net/http` + `chi`; scanner/indexador e jobs; API REST; streaming de mídia com HTTP Range.
- **Persistência:** SQLite por `modernc.org/sqlite`, FTS5, sem CGO; banco em `data/catalogo.db` e miniaturas em `data/thumbs/`, relativos à pasta do executável; `journal_mode=DELETE`, `synchronous=FULL`.
- **Frontend:** React, Vite, TypeScript e Tailwind; TanStack Query e Zustand; build embutido com `go:embed`.
- **Leitores/mídia:** `<video>` e hls.js se necessário; PDF.js; epub.js ou foliate-js; ffmpeg/ffprobe para metadados, miniaturas e conversão de vídeo sob demanda para garantir reprodução de HEVC/MKV.
- **Metadados:** goexif para EXIF; dhowden/tag para áudio; EPUB via OPF.
- **Distribuição:** GoReleaser; app.exe, `data/` e ffmpeg/; o usuário escolhe a raiz da biblioteca, enquanto os dados do aplicativo ficam em `data/` ao lado do executável.

### Entidades principais

- `items`: caminho relativo único, nome, extensão, tipo, tamanho, mtime, hash parcial, duração/dimensões/páginas, metadados, miniatura e estado removido.
- `collections`: cursos, séries, álbuns, estantes ou coleções manuais; título, descrição, capa, raiz e modo automático.
- `collection_items`: relação item/coleção com módulo e posição.
- `progress`: posição, locator CFI, percentual, concluído e data de atualização.
- `notes`: notas, destaques e marcadores com posição/locator, texto e cor.
- `tags`, `item_tags`, `favorites`, `study_log` e `settings` dão suporte à organização e estudo.
- Tabela virtual FTS5 `search` indexa nome, título, autor, conteúdo e ID do item.

## 10. API inicial

| Método e rota | Finalidade |
| --- | --- |
| `GET /api/home` | Conteúdo dos carrosséis da home |
| `GET /api/items?type=&q=&tag=&page=` | Lista paginada |
| `GET /api/items/{id}` | Detalhe e metadados |
| `GET /media/{id}` e `GET /media/{id}/transcoded` | Streaming com suporte a Range |
| `GET /thumb/{id}?size=` | Miniatura |
| `GET /api/courses` e `GET /api/courses/{id}` | Cursos, módulos e aulas |
| `GET/PUT /api/progress/{itemId}` | Consultar e salvar progresso |
| `GET/POST/DELETE /api/notes` | Consultar e gerenciar notas |
| `POST/DELETE /api/favorites/{id}` | Favoritos |
| `GET/POST /api/tags` | Tags |
| `GET /api/search?q=` | Busca global |
| `POST /api/scan` e `GET /api/scan/events` | Iniciar varredura e acompanhar SSE |
| `GET /api/stats` | Estatísticas |
| `GET /healthz` | Estado do serviço |

## 11. Segurança e observabilidade

- Bind estrito em `127.0.0.1`, porta aleatória, token de sessão no primeiro acesso e validação anti *path traversal* em todo acesso a arquivo.
- Lock file impede duas instâncias no mesmo SSD.
- Logs estruturados com rotação em `data/logs/app.log`; registrar duração e taxa da varredura, fila de miniaturas e erros do ffmpeg.
- `/healthz` e `/debug/stats` (somente local); tela de diagnóstico com versão, tamanho do banco, itens por tipo e erros recentes.

## 12. Riscos conhecidos

Corrupção do banco por remoção do SSD; formatos HEVC/MKV incompatíveis com o navegador; varredura lenta em grandes volumes; nomes irregulares em cursos; escopo amplo; PDFs escaneados sem camada de texto. Mitigações incluem journal seguro, backups e reconstrução, transcodificação em cache, varredura incremental e workers, ordenação natural com `curso.json`, entrega por fases e OCR opcional futuro.
