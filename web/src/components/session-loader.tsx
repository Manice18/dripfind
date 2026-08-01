export function SessionLoader({ label = "hold up" }: { label?: string }) {
  return (
    <main className="session-loader" aria-busy="true">
      <div className="session-loader-inner" role="status">
        <span className="session-spinner" aria-hidden="true" />
        <p className="session-loader-label">{label}</p>
      </div>
    </main>
  );
}
