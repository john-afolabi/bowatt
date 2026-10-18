export type FeedImage = {
  id: string;
  title: string;
  tags: string[];
  url: string;
  width: number;
  height: number;
  sizeBytes: number;
  createdAt: string;
};

export type ImagePage = {
  items: FeedImage[];
  nextCursor?: string;
};

export type ServerEvent = { type: "image.created"; data: FeedImage };
