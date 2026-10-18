import type { FeedImage, ImagePage, ServerEvent } from "@/types";
import type { InfiniteData, QueryClient } from "@tanstack/react-query";

export function uniqueTags(images: FeedImage[]): string[] {
  const tags = new Set<string>();
  for (const image of images) {
    for (const tag of image.tags) tags.add(tag);
  }
  return [...tags].sort();
}

export function prependImage(images: FeedImage[], incoming: FeedImage): FeedImage[] {
    if (images.some((image) => image.id === incoming.id)) return images;
    return [incoming, ...images];
  }
  
  export function prependToPages(
    data: InfiniteData<ImagePage> | undefined,
    image: FeedImage,
  ): InfiniteData<ImagePage> | undefined {
    if (!data || data.pages.length === 0) return data;
    const [first, ...rest] = data.pages;
    return { ...data, pages: [{ ...first, items: prependImage(first.items, image) }, ...rest] };
  }
  

  export function insertLive(queryClient: QueryClient, image: FeedImage): void {
    queryClient.setQueriesData(
      {
        queryKey: ["images"],
        predicate: (query) => matchesTags(image, query.queryKey[1] as string[]),
      },
      (data: InfiniteData<ImagePage> | undefined) => prependToPages(data, image),
    );
  }
  

  export function matchesTags(image: FeedImage, activeTags: string[]): boolean {
    return activeTags.every((tag) => image.tags.includes(tag));
  }

  export function isFeedImage(value: unknown): value is FeedImage {
    if (typeof value !== "object" || value === null) return false;
    const candidate = value as Record<string, unknown>;
  
    return (
      typeof candidate.id === "string" &&
      typeof candidate.title === "string" &&
      typeof candidate.url === "string" &&
      typeof candidate.createdAt === "string" &&
      typeof candidate.width === "number" &&
      typeof candidate.height === "number" &&
      typeof candidate.sizeBytes === "number" &&
      Array.isArray(candidate.tags) &&
      candidate.tags.every((tag) => typeof tag === "string")
    );
  }
  

  export function parseServerEvent(payload: unknown): ServerEvent | null {
    if (typeof payload !== "string") return null;
  
    let parsed: unknown;
    try {
      parsed = JSON.parse(payload);
    } catch {
      return null;
    }
  
    if (typeof parsed !== "object" || parsed === null) return null;
    const { type, data } = parsed as { type?: unknown; data?: unknown };
  
    if (type !== "image.created" || !isFeedImage(data)) return null;
    return { type: "image.created", data };
  }