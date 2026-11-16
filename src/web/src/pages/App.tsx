import { useState } from 'react';
import { Button, Card, Carousel, Modal } from '../ui';

export function App() {
  const [welcomeOpen, setWelcomeOpen] = useState(false);

  return (
    <main className="app-shell">
      <header className="topbar">
        <a className="brand" href="/" aria-label="Biblioteca do SSD, início">
          <span className="brand-mark" aria-hidden="true">b.</span>
          <span>Biblioteca<span className="brand-muted"> do SSD</span></span>
        </a>
        <div className="local-badge"><span /> Neste dispositivo</div>
      </header>

      <section className="welcome" aria-labelledby="welcome-title">
        <div className="welcome-copy">
          <p className="eyebrow">SEU ESPAÇO DE ESTUDO</p>
          <h1 id="welcome-title">Tudo o que você<br /><span>quer aprender.</span></h1>
          <p className="welcome-description">Uma biblioteca tranquila para seus cursos, leituras e descobertas. Seus arquivos ficam sempre com você.</p>
          <Button onClick={() => setWelcomeOpen(true)}>Conheça sua biblioteca <span aria-hidden="true">↗</span></Button>
        </div>
        <div className="welcome-art" aria-hidden="true">
          <div className="orbit orbit-one" /><div className="orbit orbit-two" />
          <div className="book-stack"><div className="book book-back" /><div className="book book-mid" /><div className="book book-front"><span>APRENDER<br />COM CALMA</span></div></div>
          <span className="sparkle sparkle-one">✳</span><span className="sparkle sparkle-two">✦</span>
        </div>
      </section>

      <section className="library-section" aria-label="Sua biblioteca">
        <Carousel title="Sua biblioteca" subtitle="Um lugar para cada ideia" action="Ver tudo">
          <Card icon="▤" title="Cursos" detail="Aulas e materiais" tone="sage" />
          <Card icon="▧" title="Leituras" detail="PDFs e livros" tone="clay" />
          <Card icon="▷" title="Vídeos" detail="Aprenda no seu ritmo" tone="blue" />
          <Card icon="✧" title="Fotos e arquivos" detail="Suas referências" tone="sand" />
        </Carousel>
      </section>
      <footer><span>Feito para aprender no seu tempo.</span><span>Seus dados ficam no seu SSD <span className="lock">▣</span></span></footer>

      <Modal open={welcomeOpen} title="Sua biblioteca começa aqui" onClose={() => setWelcomeOpen(false)}>
        <p>A base está pronta. Na próxima etapa, você poderá escolher uma pasta do SSD e organizar seus materiais de estudo.</p>
        <Button onClick={() => setWelcomeOpen(false)}>Entendi</Button>
      </Modal>
    </main>
  );
}
