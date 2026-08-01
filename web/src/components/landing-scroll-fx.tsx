"use client";

import { useEffect } from "react";

const SECTION_SEL = "[data-landing-section]";
const LINK_SEL = '.landing-jump a[href^="#"]';

function prefersReducedMotion() {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

export function LandingScrollFX() {
  useEffect(() => {
    const root = document.documentElement;
    const reduced = prefersReducedMotion();

    const sections = Array.from(
      document.querySelectorAll<HTMLElement>(SECTION_SEL)
    );
    const links = Array.from(
      document.querySelectorAll<HTMLAnchorElement>(LINK_SEL)
    );

    const setActive = (id: string) => {
      for (const link of links) {
        const match = link.getAttribute("href") === `#${id}`;
        link.classList.toggle("is-active", match);
        if (match) {
          link.setAttribute("aria-current", "true");
        } else {
          link.removeAttribute("aria-current");
        }
      }
    };

    // Mark in-view sections before arming motion CSS so first paint stays readable.
    for (const section of sections) {
      if (reduced) {
        section.classList.add("is-visible");
        continue;
      }
      const rect = section.getBoundingClientRect();
      if (rect.top < window.innerHeight * 0.85 && rect.bottom > 0) {
        section.classList.add("is-visible");
      }
    }

    if (!reduced) {
      root.classList.add("js-landing-motion");
    }

    const revealObs = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            entry.target.classList.add("is-visible");
            revealObs.unobserve(entry.target);
          }
        }
      },
      { threshold: 0.16, rootMargin: "0px 0px -10% 0px" }
    );

    const navObs = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio);
        const top = visible[0]?.target;
        if (top?.id) setActive(top.id);
      },
      { threshold: [0.25, 0.45, 0.6], rootMargin: "-18% 0px -45% 0px" }
    );

    for (const section of sections) {
      if (!reduced && !section.classList.contains("is-visible")) {
        revealObs.observe(section);
      }
      navObs.observe(section);
    }

    const onClick = (event: Event) => {
      const link = event.currentTarget as HTMLAnchorElement;
      const href = link.getAttribute("href");
      if (!href?.startsWith("#")) return;
      const target = document.querySelector<HTMLElement>(href);
      if (!target) return;
      event.preventDefault();
      target.scrollIntoView({
        behavior: reduced ? "auto" : "smooth",
        block: "start",
      });
      setActive(href.slice(1));
      window.setTimeout(() => {
        target.focus({ preventScroll: true });
      }, reduced ? 0 : 420);
    };

    for (const link of links) {
      link.addEventListener("click", onClick);
    }

    return () => {
      root.classList.remove("js-landing-motion");
      revealObs.disconnect();
      navObs.disconnect();
      for (const link of links) {
        link.removeEventListener("click", onClick);
      }
    };
  }, []);

  return null;
}
