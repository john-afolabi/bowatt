import type { FeedImage } from "@/types";

export function uniqueTags(images: FeedImage[]): string[] {
  const tags = new Set<string>();
  for (const image of images) {
    for (const tag of image.tags) tags.add(tag);
  }
  return [...tags].sort();
}
