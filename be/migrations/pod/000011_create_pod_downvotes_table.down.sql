-- Migration: Remove pod_downvotes table

DROP INDEX IF EXISTS idx_pod_downvotes_user_id;
DROP INDEX IF EXISTS idx_pod_downvotes_pod_id;
DROP TABLE IF EXISTS pod_downvotes;
