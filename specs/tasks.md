# Tarefas SDD — Biblioteca do SSD

As caixas marcam o acompanhamento do trabalho. A sequência respeita as dependências entre fundação, catálogo e experiências de consumo. As estimativas são as do plano original.

## Fase 0 — Fundação

- [x] Criar repositório e configurar CI para build e testes.
- [x] Criar esqueleto do backend Go e frontend Vite/React/TypeScript.
- [x] Configurar build de produção e embutir frontend com `go:embed`.
- [x] Implementar resolução da raiz da biblioteca escolhida e usar `data/` ao lado do executável para os dados do aplicativo.
- [x] Configurar SQLite via `modernc.org/sqlite`, migrations iniciais, `journal_mode=DELETE` e `synchronous=FULL`.
- [x] Adicionar configuração e logs estruturados com `slog`.
- [x] Implementar `/healthz` e inicialização do app abrindo a home local.
- [x] Criar tokens visuais, tipografia e componentes base (Card, Carousel, Button, Modal).
- [x] Confirmar: `.exe` abre a página inicial e cria o banco no SSD.

## Fase 1 — Catálogo e navegação

- [x] Implementar classificação de extensões nos tipos photo/video/audio/pdf/ebook/other.
- [x] Implementar varredura recursiva com exclusões configuráveis e de sistema.
- [x] Tornar a varredura incremental por tamanho e mtime; marcar ausentes como removidos.
- [x] Calcular hash parcial para suporte à identificação de duplicados.
- [x] Criar pipeline de workers e eventos de progresso via SSE.
- [x] Implementar endpoints de scan, eventos, listagem paginada e detalhe de item.
- [x] Construir tela de explorar pastas e listagens por tipo.
- [x] Confirmar: varrer 50 mil arquivos em minutos e reabrir sem reprocessar tudo.

## Fase 2 — Fotos e PDFs

- [ ] Gerar miniaturas de fotos em 320 px e 1080 px; prever HEIC via ffmpeg.
- [ ] Extrair EXIF e ordenar fotos por data em grade/timeline.
- [ ] Implementar visualizador de fotos com zoom, tela cheia e setas.
- [ ] Gerar capas e contagem de páginas de PDF.
- [ ] Integrar PDF.js com índice, busca, progresso e marcadores persistidos.
- [ ] Persistir destaques e anotações feitos em PDFs.
- [ ] Confirmar: foto e PDF abrem; PDF reabre na página anterior.

## Fase 3 — Vídeo e cursos (MVP)

- [ ] Integrar ffprobe para duração, codec, capítulos, trilhas de áudio e legendas.
- [ ] Gerar miniatura de vídeo e sprite de preview quando aplicável.
- [ ] Implementar streaming local com HTTP Range.
- [ ] Construir player com velocidade, legendas, atalhos, tela cheia, PiP e retomada.
- [ ] Detectar cursos, módulos, aulas, capas, materiais anexos e legendas associadas.
- [ ] Aplicar ordenação natural e ler substituições opcionais de `curso.json`.
- [ ] Implementar listagem/detalhe de cursos e autoplay da próxima aula.
- [ ] Criar home com banner e carrosséis “Continuar” e “Meus cursos”.
- [ ] Adicionar transcodificação sob demanda para formatos incompatíveis com cache.
- [ ] Confirmar: curso real do SSD toca até o fim e mantém o progresso.

## Fase 4 — Livros e EPUB

- [ ] Extrair capa, título e autor de EPUB via OPF.
- [ ] Integrar leitor EPUB com CFI, índice, busca e controle de fonte/entrelinha/margens.
- [ ] Adicionar temas claro, escuro e sépia.
- [ ] Implementar estante e estados não lido/lendo/lido.
- [ ] Persistir notas, destaques coloridos, marcadores e percentual de leitura.
- [ ] Confirmar: EPUB e PDF mantêm progresso e destaques após reabrir.

## Fase 5 — Música, busca e organização

- [ ] Extrair tags ID3 e capa de áudio.
- [ ] Implementar agrupamento por álbum/artista, fila, playlists e mini-player persistente.
- [ ] Indexar nome, título, autor e texto extraído de PDFs/EPUBs com FTS5.
- [ ] Definir e aplicar um limite fixo para indexação do conteúdo; informar quais arquivos excederem esse limite.
- [ ] Implementar busca global agrupada por tipo, filtros, relevância e snippets.
- [ ] Adicionar tags, favoritos e coleções manuais.
- [ ] Confirmar: busca encontra termo existente dentro de PDF.

## Fase 6 — Estudo e polimento

- [ ] Implementar notas com timestamp em vídeos.
- [ ] Registrar tempo/páginas por dia e apresentar metas, histórico, estatísticas e streak.
- [ ] Permitir configurar o limiar de conclusão de vídeos (padrão de 90%) e reverter uma conclusão.
- [ ] Expor detecção de duplicados.
- [ ] Implementar backup rotativo do catálogo e tela de configurações/reindexação.
- [ ] Implementar checagem de integridade do SQLite na abertura.
- [ ] Adicionar acessibilidade de teclado, contraste, foco e atributos ARIA.
- [ ] Virtualizar listas e tratar erros de mídia, banco e SSD removido.
- [ ] Implementar `/debug/stats` local e tela de diagnóstico.
- [ ] Confirmar: uma semana de uso sem travar ou perder progresso.

## Fase 7 — Empacotamento e release

- [ ] Configurar build Windows amd64 sem CGO e otimização do binário.
- [ ] Configurar GoReleaser e empacotar frontend, app.exe, ffmpeg/ffprobe e estrutura data/.
- [ ] Criar ícone e script/atalho “Abrir Biblioteca”.
- [ ] Escrever guia de uso e checklist de restauração do catálogo.
- [ ] Confirmar: pasta de release pode ser copiada ao SSD e aberta em Windows 10/11.

## Itens para validar antes do release

- [ ] Preparar pasta de amostra com um curso, alguns livros e fotos.
- [ ] Prototipar a home usando a paleta e os componentes definidos.
- [ ] Validar operação no SSD NTFS, incluindo alteração de letra da unidade e remoção segura.

## Testes planejados

- [ ] Unitários: ordenação natural, detecção de cursos, classificação, proteção anti *path traversal*.
- [ ] Integração: scanner com pasta fixture e resultado esperado.
- [ ] E2E: home → curso → player → retomada de progresso.
- [ ] Desempenho: benchmark de arquivos por segundo e latência da API com 100 mil itens.
- [ ] Resiliência: interromper processo durante escrita e verificar `PRAGMA integrity_check`.
