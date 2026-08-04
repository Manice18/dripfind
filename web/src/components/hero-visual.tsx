import Image from "next/image";

export function HeroVisual({ priority = false }: { priority?: boolean }) {
  return (
    <div className="hero-visual" aria-hidden="true">
      <div className="hero-photo">
        <Image
          src="/heroes/dripfind-hero.jpg"
          alt=""
          fill
          priority={priority}
          sizes="100vw"
          quality={70}
          className="hero-photo-img"
        />
      </div>
      <div className="hero-wash" />
      <div className="hero-grain" />
    </div>
  );
}
