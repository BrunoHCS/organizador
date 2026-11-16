# Constituição do Projeto — Biblioteca do SSD

Esta constituição define os princípios que orientam as decisões de produto e implementação. Em caso de conflito, os princípios abaixo prevalecem; mudanças devem ser registradas nesta constituição e refletidas nos documentos relacionados.

## 1. Offline, local e privado

- O aplicativo funciona sem conta, nuvem ou conexão com a internet.
- O servidor local escuta somente em `127.0.0.1`; não há telemetria.
- Catálogo, progresso, notas, favoritos e configurações permanecem no SSD.
- Fontes, interface e dependências necessárias à experiência offline não dependem de CDN.

## 2. Portabilidade em primeiro lugar

- O produto deve rodar em Windows 10/11 sem instalador, a partir de um `.exe` portátil e dos arquivos auxiliares necessários.
- O aplicativo descobre sua raiz a partir do local do executável e usa caminhos relativos, tolerando mudança da letra da unidade.
- O backend é compilado sem CGO; o frontend de produção é embutido no executável.
- Dependências externas de mídia, como ffmpeg/ffprobe, são distribuídas junto ao aplicativo quando aplicável.

## 3. Integridade e recuperação dos dados

- A remoção inesperada do SSD não deve corromper o catálogo de forma silenciosa.
- SQLite usa `journal_mode=DELETE` e `synchronous=FULL`; o banco é fechado corretamente no encerramento.
- Manter backups rotativos do catálogo, verificar integridade na abertura e permitir reconstrução por nova varredura.
- O catálogo deve registrar itens ausentes sem apagar imediatamente o histórico associado.

## 4. Acesso seguro ao sistema de arquivos

- Acesso a arquivos é limitado à raiz configurada do aplicativo/SSD e validado contra *path traversal*.
- A API local usa token de sessão no primeiro acesso e porta aleatória.
- Uma instância por SSD é garantida por lock file.
- Pastas de sistema e a própria pasta do aplicativo são excluídas da varredura por padrão.

## 5. Experiência voltada ao estudo

- A interface combina navegação familiar por carrosséis e cards com linguagem visual calma, focada e acolhedora.
- Progresso, retomada, cursos e leitura têm prioridade sobre elementos puramente decorativos.
- A experiência deve oferecer navegação completa por teclado, estados de carregamento e feedback claro de erro.
- O tema escuro é padrão; tema claro e modo sépia no leitor são opções.

## 6. Desempenho em bibliotecas grandes

- O catálogo e a home devem continuar responsivos com até 100 mil itens indexados.
- A varredura é incremental, evita reprocessar arquivos sem alteração e usa workers limitados.
- Listas extensas usam paginação ou virtualização; miniaturas são geradas e carregadas sob demanda.
- A meta da home é abrir em menos de 1 segundo e a rolagem atingir 60 fps em hardware compatível.

## 7. Entrega incremental e escopo controlado

- Implementar em fases, entregando primeiro a fundação e o MVP de catálogo, fotos/PDFs, vídeo e cursos.
- Cada fase deve ter um resultado verificável descrito em `plan.md` e tarefas rastreáveis em `tasks.md`.
- Funcionalidades de baixa prioridade e itens de backlog não podem bloquear o MVP.
- Multiusuário, sincronização em nuvem, aplicativo mobile e DRM estão fora do escopo atual.

## 8. Arquitetura simples e compatível com a distribuição

- Backend em Go (`net/http` e `chi`), SQLite sem CGO e frontend React + TypeScript + Vite + Tailwind.
- A API REST serve também mídia com HTTP Range; progresso de varredura pode ser transmitido por SSE.
- Evitar serviços externos obrigatórios em tempo de execução.
- Evoluções de esquema são aplicadas por migrations e compatíveis com abertura do catálogo existente.

## 9. Decisões e alterações

- Requisitos e critérios de aceite vivem em `spec.md`; estratégia e fases em `plan.md`; trabalho executável em `tasks.md`.
- Alterações de escopo, arquitetura, privacidade ou portabilidade devem atualizar os documentos SDD correspondentes.
- A fonte inicial destes princípios é `Plano de Implementação.md`.
