import { useMemo } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";

import { describeError, listImages } from "@/lib/api";
import type { FeedImage } from "@/types";
import { uniqueTags } from "@/lib/feed";

export const PAGE_SIZE = 24;

export type FeedStatus = "loading" | "ready" | "error";

export type Feed = {
  images: FeedImage[];
  tags: string[];
  status: FeedStatus;
  error?: string;
  moreError?: string;
  loadingMore: boolean;
  hasMore: boolean;
  loadMore: () => void;
  reload: () => void;
};

export function useFeed(activeTags: string[] = []): Feed {
  const tagsKey = [...activeTags].sort();

  const query = useInfiniteQuery({
    queryKey: ["images", tagsKey],
    queryFn: ({ pageParam, signal }) =>
      listImages({ limit: PAGE_SIZE, cursor: pageParam, tags: tagsKey }, signal),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.nextCursor,
  });

  const images = useMemo(() => query.data?.pages.flatMap((page) => page.items) ?? [], [query.data]);
  const tags = useMemo(() => uniqueTags(images), [images]);

  return {
    images,
    tags,
    status: query.data !== undefined ? "ready" : query.isError ? "error" : "loading",
    error: query.error ? describeError(query.error) : undefined,
    moreError:
      query.isFetchNextPageError && !query.isFetchingNextPage
        ? describeError(query.error)
        : undefined,
    loadingMore: query.isFetchingNextPage,
    hasMore: query.hasNextPage,
    loadMore: () => void query.fetchNextPage({ cancelRefetch: false }),
    reload: () => void query.refetch(),
  };
}
