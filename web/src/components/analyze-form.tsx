"use client";

import { FormEvent, useCallback, useRef, useState } from "react";

const ACCEPT =
  "image/jpeg,image/png,image/webp,image/gif,.jpg,.jpeg,.png,.webp,.gif";
const MAX_BYTES = 12 * 1024 * 1024;

export function AnalyzeForm({
  onSubmitUrl,
  onSubmitFile,
  loading,
}: {
  onSubmitUrl: (url: string) => void;
  onSubmitFile: (file: File) => void;
  loading?: boolean;
}) {
  const [url, setUrl] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [dragOver, setDragOver] = useState(false);
  const [localError, setLocalError] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const clearFile = useCallback(() => {
    if (preview) URL.revokeObjectURL(preview);
    setFile(null);
    setPreview(null);
    if (inputRef.current) inputRef.current.value = "";
  }, [preview]);

  const pickFile = useCallback(
    (next: File | null) => {
      setLocalError("");
      if (!next) {
        clearFile();
        return;
      }
      if (!next.type.startsWith("image/")) {
        setLocalError("Please choose an image file (JPEG, PNG, WebP, or GIF).");
        return;
      }
      if (next.size > MAX_BYTES) {
        setLocalError("Image is too large (max 12MB).");
        return;
      }
      if (preview) URL.revokeObjectURL(preview);
      setFile(next);
      setPreview(URL.createObjectURL(next));
    },
    [clearFile, preview],
  );

  function handleUrlSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = url.trim();
    if (!trimmed || loading) return;
    onSubmitUrl(trimmed);
  }

  function handleUpload() {
    if (!file || loading) return;
    onSubmitFile(file);
  }

  return (
    <div className="analyze-stack">
      <form
        onSubmit={handleUrlSubmit}
        className="analyze-path analyze-path-url"
      >
        <header className="analyze-path-head">
          <p className="analyze-path-label">Pinterest</p>
          <p className="analyze-path-hint">Paste a pin URL</p>
        </header>
        <div className="analyze-url-row">
          <label htmlFor="outfit-url" className="sr-only">
            Pinterest URL
          </label>
          <input
            id="outfit-url"
            className="analyze-url-input"
            type="url"
            required
            placeholder="https://pin.it/… or pinterest.com/pin/…"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            disabled={loading}
          />
          <button
            type="submit"
            className="analyze-cta"
            disabled={loading || !url.trim()}
          >
            {loading ? "Reading look…" : "Recreate"}
          </button>
        </div>
      </form>

      <div className="analyze-divider" aria-hidden>
        <span>or</span>
      </div>

      <div className="analyze-path analyze-path-photo">
        <header className="analyze-path-head">
          <p className="analyze-path-label">Photo</p>
          <p className="analyze-path-hint">
            Drop, browse, or paste a screenshot
          </p>
        </header>
        <div
          className={`dropzone${dragOver ? " is-dragover" : ""}${preview ? " has-preview" : ""}`}
          onDragEnter={(e) => {
            e.preventDefault();
            setDragOver(true);
          }}
          onDragOver={(e) => {
            e.preventDefault();
            setDragOver(true);
          }}
          onDragLeave={(e) => {
            e.preventDefault();
            setDragOver(false);
          }}
          onDrop={(e) => {
            e.preventDefault();
            setDragOver(false);
            const dropped = e.dataTransfer.files?.[0];
            if (dropped) pickFile(dropped);
          }}
          onPaste={(e) => {
            const items = e.clipboardData?.items;
            if (!items) return;
            for (const item of items) {
              if (item.type.startsWith("image/")) {
                const pasted = item.getAsFile();
                if (pasted) {
                  e.preventDefault();
                  pickFile(pasted);
                  break;
                }
              }
            }
          }}
          tabIndex={0}
          role="button"
          aria-label="Upload outfit photo"
          aria-disabled={loading || undefined}
          onClick={() => inputRef.current?.click()}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              inputRef.current?.click();
            }
          }}
        >
          <input
            ref={inputRef}
            type="file"
            accept={ACCEPT}
            className="sr-only"
            disabled={loading}
            onChange={(e) => pickFile(e.target.files?.[0] ?? null)}
          />
          {preview ? (
            <div className="dropzone-preview">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={preview} alt="Selected outfit preview" />
              <div className="dropzone-preview-meta">
                <span>{file?.name}</span>
                <button
                  type="button"
                  className="dropzone-clear"
                  onClick={(e) => {
                    e.stopPropagation();
                    clearFile();
                  }}
                >
                  Remove
                </button>
              </div>
            </div>
          ) : (
            <div className="dropzone-empty">
              <strong>Drop a photo here</strong>
              <span>or click to browse</span>
            </div>
          )}
        </div>
        {localError && (
          <p className="error analyze-path-error" role="alert">
            {localError}
          </p>
        )}
        <button
          type="button"
          className="analyze-cta analyze-cta-photo"
          disabled={loading || !file}
          onClick={handleUpload}
        >
          {loading ? "Reading look…" : "Recreate from photo"}
        </button>
      </div>
    </div>
  );
}
