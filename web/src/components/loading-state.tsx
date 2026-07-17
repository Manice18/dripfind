"use client";

const STAGES = [
  "Pulling the image",
  "Reading the silhouette",
  "Naming every piece",
  "Searching the racks",
  "Ranking closest matches",
];

export function LoadingState({ stage = 0 }: { stage?: number }) {
  const active = Math.min(stage, STAGES.length - 1);

  return (
    <section className="loading-panel" aria-live="polite">
      <div className="loading-orbit" aria-hidden />
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
