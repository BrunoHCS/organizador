import type { ReactNode } from 'react';

export function Button({ children, onClick, variant = 'primary' }: { children: ReactNode; onClick?: () => void; variant?: 'primary' | 'quiet' }) {
  return <button className={`button button-${variant}`} onClick={onClick}>{children}</button>;
}

export function Card({ icon, title, detail, tone }: { icon: string; title: string; detail: string; tone: string }) {
  return <button className="library-card" type="button"><span className={`card-icon ${tone}`} aria-hidden="true">{icon}</span><span className="card-copy"><strong>{title}</strong><small>{detail}</small></span><span className="card-arrow" aria-hidden="true">↗</span></button>;
}

export function Carousel({ title, subtitle, action, children }: { title: string; subtitle: string; action: string; children: ReactNode }) {
  return <div className="carousel"><div className="section-heading"><div><p className="eyebrow">{subtitle}</p><h2>{title}</h2></div><button className="text-action" type="button">{action} <span aria-hidden="true">→</span></button></div><div className="card-row">{children}</div></div>;
}

export function Modal({ open, title, onClose, children }: { open: boolean; title: string; onClose: () => void; children: ReactNode }) {
  if (!open) return null;
  return <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}><section className="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title"><button className="modal-close" onClick={onClose} aria-label="Fechar">×</button><p className="eyebrow">BIBLIOTECA DO SSD</p><h2 id="modal-title">{title}</h2><div className="modal-content">{children}</div></section></div>;
}
