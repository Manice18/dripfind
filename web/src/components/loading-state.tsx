"use client";

import { useEffect, useState } from "react";

const STAGES = [
  "Pulling the image",
  "Reading the silhouette",
  "Naming every piece",
  "Searching the racks",
  "Ranking closest matches",
];

export function LoadingState() {
  const [stage, setStage] = useState(0);
  const active = Math.min(stage, STAGES.length - 1);

  useEffect(() => {
    const t = setInterval(
      () => setStage((s) => (s + 1) % STAGES.length),
      2200,
    );
    return () => clearInterval(t);
  }, []);

  return (
    <section className="loading-panel" aria-live="polite">
      <div className="loading-orbit-wrap" aria-hidden>
        <div className="loading-orbit" />
      </div>
      <p className="eyebrow">Working</p>
      <h2>{STAGES[active]}</h2>
      <ol className="stage-list">
        {STAGES.map((label, i) => (
          <li key={label} className={i <= active ? "done" : ""}>
            <span />
            {label}
          </li>
        ))}
      </ol>
    </section>
  );
}
