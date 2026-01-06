import { starPod, unstarPod } from "../api/pod";

export async function toggleStarPod(podId: string, isStarred: boolean): Promise<void> {
  try {
    if (isStarred) {
      await unstarPod(podId);
    } else {
      await starPod(podId);
    }
  } catch (error) {
    console.error("Error toggling star pod:", error);
    throw error;
  }
}