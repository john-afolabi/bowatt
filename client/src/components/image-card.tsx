import type { FeedImage } from "@/types";
import { Badge } from "./ui/badge";
import { absoluteTime, formatBytes } from "@/lib/format";
import { relativeTime } from "@/lib/format";

type Props = {
  image: FeedImage;
  activeTags: string[];
  onToggleTag: (tag: string) => void;
  onOpen: (image: FeedImage) => void;
};

export function ImageCard({ image, activeTags, onToggleTag, onOpen }: Props) {
  return (
    <article className="group overflow-hidden rounded-xl border bg-card shadow-sm transition-shadow hover:shadow-md">
      <button
        type="button"
        onClick={() => onOpen(image)}
        aria-label={`Open ${image.title}`}
        className="block w-full cursor-zoom-in overflow-hidden bg-muted"
      >
        <img
          src={image.url}
          alt={image.title}
          width={image.width}
          height={image.height}
          loading="lazy"
          decoding="async"
          className="aspect-4/3 w-full object-cover transition-transform duration-300 group-hover:scale-[1.02]"
        />
      </button>

      <div className="space-y-2 p-3">
        <h2 className="line-clamp-2 text-sm leading-snug font-medium">{image.title}</h2>
        {image.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {image.tags.map((tag) => {
              const active = activeTags.includes(tag);
              return (
                <button
                  key={tag}
                  type="button"
                  onClick={() => onToggleTag(tag)}
                  aria-pressed={active}
                  className="rounded-full focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:outline-none"
                >
                  <Badge variant={active ? "default" : "secondary"} className="cursor-pointer">
                    {tag}
                  </Badge>
                </button>
              );
            })}
          </div>
        )}

        <p className="flex flex-wrap items-center gap-x-1.5 gap-y-0.5 text-xs text-muted-foreground">
          <span title={absoluteTime(image.createdAt)}>{relativeTime(image.createdAt)}</span>
          <span aria-hidden>·</span>
          <span>
            {image.width}×{image.height}
          </span>
          <span aria-hidden>·</span>
          <span>{formatBytes(image.sizeBytes)}</span>
        </p>
      </div>
    </article>
  );
}
