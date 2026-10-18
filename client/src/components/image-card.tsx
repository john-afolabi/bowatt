import type { FeedImage } from "@/types";

type Props = {
  image: FeedImage;
};

export function ImageCard({ image }: Props) {
  return (
    <article className="overflow-hidden rounded-xl border bg-card shadow-sm">
      <img
        src={image.url}
        alt={image.title}
        width={image.width}
        height={image.height}
        loading="lazy"
        decoding="async"
        className="aspect-4/3 w-full object-cover"
      />

      <div className="space-y-2 p-3">
        <h2 className="line-clamp-2 text-sm leading-snug font-medium">{image.title}</h2>
        <p className="flex flex-wrap items-center gap-x-1.5 gap-y-0.5 text-xs text-muted-foreground">
          <span>
            {image.width}×{image.height}
          </span>
        </p>
      </div>
    </article>
  );
}
