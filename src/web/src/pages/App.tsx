import { useCallback, useEffect, useState } from 'react';
import { Button } from '../ui';

type Item = { id: number; path: string; name: string; extension: string; type: string; size: number; modifiedAt: string; partialHash: string; removed: boolean };
type ScanProgress = { visited: number; indexed: number; errors: number; done: boolean; message?: string };
const categories = [{ id: 'all', label: 'Tudo' }, { id: 'photo', label: 'Fotos' }, { id: 'video', label: 'Vídeos' }, { id: 'audio', label: 'Áudios' }, { id: 'pdf', label: 'PDFs' }, { id: 'ebook', label: 'E-books' }, { id: 'other', label: 'Outros' }];
const sizeLabel = (size: number) => size < 1024 ? `${size} B` : size < 1024 ** 2 ? `${(size / 1024).toFixed(1)} KB` : `${(size / 1024 ** 2).toFixed(1)} MB`;

export function App() {
  const [items, setItems] = useState<Item[]>([]); const [folders, setFolders] = useState<string[]>([]); const [folder, setFolder] = useState('');
  const [kind, setKind] = useState('all'); const [page, setPage] = useState(1); const [total, setTotal] = useState(0);
  const [progress, setProgress] = useState<ScanProgress | null>(null); const [scanning, setScanning] = useState(false); const [selected, setSelected] = useState<Item | null>(null); const [error, setError] = useState('');

  const refresh = useCallback(async () => {
    try { const params = new URLSearchParams({ page: String(page), pageSize: '50', type: kind }); if (folder) params.set('folder', folder);
      const [listing, tree] = await Promise.all([fetch(`/api/items?${params}`), fetch('/api/folders')]);
      if (!listing.ok || !tree.ok) throw new Error('Não foi possível carregar o catálogo.');
      const data = await listing.json(); const folderData = await tree.json(); setItems(data.items); setTotal(data.total); setFolders(folderData.folders); setError('');
    } catch (e) { setError(e instanceof Error ? e.message : 'Falha ao carregar o catálogo.'); }
  }, [folder, kind, page]);
  useEffect(() => { void refresh(); }, [refresh]);

  const startScan = async () => {
    setError(''); setProgress({ visited: 0, indexed: 0, errors: 0, done: false });
    try { const response = await fetch('/api/scan', { method: 'POST' }); if (!response.ok) throw new Error(response.status === 409 ? 'Uma varredura já está em andamento.' : 'Não foi possível iniciar a varredura.'); setScanning(true);
      const events = new EventSource('/api/scan/events'); events.addEventListener('progress', (event) => { const next = JSON.parse((event as MessageEvent).data) as ScanProgress; setProgress(next); setScanning(!next.done); if (next.done) { events.close(); void refresh(); } }); events.onerror = () => { events.close(); setScanning(false); };
    } catch (e) { setError(e instanceof Error ? e.message : 'Falha ao iniciar a varredura.'); }
  };
  const openItem = async (item: Item) => { try { const response = await fetch(`/api/items/${item.id}`); if (response.ok) setSelected(await response.json()); } catch { setError('Não foi possível abrir os detalhes do arquivo.'); } };

  return <main className="app-shell">
    <header className="topbar"><a className="brand" href="/" aria-label="Biblioteca do SSD, início"><span className="brand-mark" aria-hidden="true">b.</span><span>Biblioteca<span className="brand-muted"> do SSD</span></span></a><div className="local-badge"><span /> Neste dispositivo</div></header>
    <section className="welcome" aria-labelledby="welcome-title"><div className="welcome-copy"><p className="eyebrow">SEU ESPAÇO DE ESTUDO</p><h1 id="welcome-title">Tudo o que você<br /><span>quer aprender.</span></h1><p className="welcome-description">Explore os arquivos da sua biblioteca local. Seus materiais ficam sempre com você.</p><Button onClick={() => document.getElementById('explorar')?.scrollIntoView({ behavior: 'smooth' })}>Explorar biblioteca <span aria-hidden="true">↓</span></Button></div><div className="welcome-art" aria-hidden="true"><div className="orbit orbit-one" /><div className="orbit orbit-two" /><div className="book-stack"><div className="book book-back" /><div className="book book-mid" /><div className="book book-front"><span>APRENDER<br />COM CALMA</span></div></div><span className="sparkle sparkle-one">✳</span><span className="sparkle sparkle-two">✦</span></div></section>
    <section id="explorar" className="explorer" aria-labelledby="explorer-title"><div className="explorer-heading"><div><p className="eyebrow">CATÁLOGO LOCAL</p><h2 id="explorer-title">Explorar arquivos</h2><p>Pastas e arquivos indexados no SSD</p></div><Button onClick={startScan}>{scanning ? 'Varredura em andamento…' : '↻  Varrer biblioteca'}</Button></div>
      {progress && <div className="scan-status" role="status"><span>{progress.done ? progress.message : `Lendo arquivos… ${progress.visited} visitados`}</span><span>{progress.indexed} indexados{progress.errors ? ` · ${progress.errors} erros` : ''}</span></div>}
      {error && <p className="inline-error" role="alert">{error}</p>}
      <div className="browse-controls"><label htmlFor="folder-filter">Pasta</label><select id="folder-filter" value={folder} onChange={(e) => { setFolder(e.target.value); setPage(1); }}><option value="">Todas as pastas</option>{folders.filter(Boolean).map((name) => <option key={name} value={name}>{name}</option>)}</select></div>
      <div className="type-filters" role="tablist" aria-label="Filtrar por tipo">{categories.map((category) => <button key={category.id} role="tab" aria-selected={kind === category.id} className={kind === category.id ? 'filter active' : 'filter'} onClick={() => { setKind(category.id); setPage(1); }}>{category.label}</button>)}</div>
      {items.length ? <div className="item-list">{items.map((item) => <button className="item-row" key={item.id} onClick={() => void openItem(item)}><span className={`item-type ${item.type}`}>{item.type === 'photo' ? '▧' : item.type === 'video' ? '▷' : item.type === 'audio' ? '♫' : item.type === 'pdf' ? '▤' : '•'}</span><span className="item-main"><strong>{item.name}</strong><small>{item.path}</small></span><span className="item-meta">{sizeLabel(item.size)}</span><span className="item-open" aria-hidden="true">›</span></button>)}</div> : <div className="empty-catalog"><span aria-hidden="true">⌕</span><strong>Nenhum arquivo nesta seleção</strong><p>Inicie uma varredura para catalogar os arquivos da pasta configurada.</p></div>}
      <div className="pagination"><span>{total} arquivo{total === 1 ? '' : 's'}</span><div><button disabled={page <= 1} onClick={() => setPage(page - 1)}>← Anterior</button><span>Página {page} de {Math.max(1, Math.ceil(total / 50))}</span><button disabled={page * 50 >= total} onClick={() => setPage(page + 1)}>Próxima →</button></div></div>
    </section>
    <footer><span>Feito para aprender no seu tempo.</span><span>Seus dados ficam no seu SSD <span className="lock">▣</span></span></footer>
    {selected && <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setSelected(null); }}><section className="modal" role="dialog" aria-modal="true" aria-labelledby="detail-title"><button className="modal-close" onClick={() => setSelected(null)} aria-label="Fechar">×</button><p className="eyebrow">DETALHE DO ARQUIVO</p><h2 id="detail-title">{selected.name}</h2><div className="modal-content"><dl className="item-details"><dt>Caminho</dt><dd>{selected.path}</dd><dt>Tipo</dt><dd>{selected.type}</dd><dt>Tamanho</dt><dd>{sizeLabel(selected.size)}</dd><dt>Modificado</dt><dd>{new Date(selected.modifiedAt).toLocaleString()}</dd><dt>Hash parcial</dt><dd className="hash-value">{selected.partialHash}</dd></dl></div></section></div>}
  </main>;
}
