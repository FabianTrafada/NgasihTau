package qdrant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ngasihtau/internal/ai/domain"
)

type Config struct {
	Host           string
	Port           int
	CollectionName string
	VectorSize     int
}

type Client struct {
	conn              *grpc.ClientConn
	pointsClient      pb.PointsClient
	collectionsClient pb.CollectionsClient
	collectionName    string
	vectorSize        int
}

func NewClient(cfg Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
	}

	return &Client{
		conn:              conn,
		pointsClient:      pb.NewPointsClient(conn),
		collectionsClient: pb.NewCollectionsClient(conn),
		collectionName:    cfg.CollectionName,
		vectorSize:        cfg.VectorSize,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) EnsureCollection(ctx context.Context) error {
	_, err := c.collectionsClient.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: c.collectionName,
	})
	if err == nil {
		return nil
	}

	_, err = c.collectionsClient.Create(ctx, &pb.CreateCollection{
		CollectionName: c.collectionName,
		VectorsConfig: &pb.VectorsConfig{
			Config: &pb.VectorsConfig_Params{
				Params: &pb.VectorParams{
					Size:     uint64(c.vectorSize),
					Distance: pb.Distance_Cosine,
				},
			},
		},
	})

	return err
}

func (c *Client) Upsert(ctx context.Context, chunks []domain.MaterialChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	points := make([]*pb.PointStruct, len(chunks))
	for i, chunk := range chunks {
		vectors := make([]float32, len(chunk.Embedding))
		copy(vectors, chunk.Embedding)

		points[i] = &pb.PointStruct{
			Id: &pb.PointId{
				PointIdOptions: &pb.PointId_Uuid{Uuid: chunk.ID},
			},
			Vectors: &pb.Vectors{
				VectorsOptions: &pb.Vectors_Vector{
					Vector: &pb.Vector{Data: vectors},
				},
			},
			Payload: map[string]*pb.Value{
				"material_id": {Kind: &pb.Value_StringValue{StringValue: chunk.MaterialID.String()}},
				"pod_id":      {Kind: &pb.Value_StringValue{StringValue: chunk.PodID.String()}},
				"chunk_index": {Kind: &pb.Value_IntegerValue{IntegerValue: int64(chunk.ChunkIndex)}},
				"text":        {Kind: &pb.Value_StringValue{StringValue: chunk.Text}},
			},
		}
	}

	_, err := c.pointsClient.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: c.collectionName,
		Points:         points,
	})

	if err != nil {
		return fmt.Errorf("qdrant upsert failed: %w", err)
	}

	return nil
}

func (c *Client) Search(ctx context.Context, embedding []float32, materialID *uuid.UUID, podID *uuid.UUID, limit int) ([]domain.MaterialChunk, error) {
	var filter *pb.Filter
	if materialID != nil {
		filter = &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "material_id",
							Match: &pb.Match{
								MatchValue: &pb.Match_Keyword{Keyword: materialID.String()},
							},
						},
					},
				},
			},
		}
	} else if podID != nil {
		filter = &pb.Filter{
			Must: []*pb.Condition{
				{
					ConditionOneOf: &pb.Condition_Field{
						Field: &pb.FieldCondition{
							Key: "pod_id",
							Match: &pb.Match{
								MatchValue: &pb.Match_Keyword{Keyword: podID.String()},
							},
						},
					},
				},
			},
		}
	}

	resp, err := c.pointsClient.Search(ctx, &pb.SearchPoints{
		CollectionName: c.collectionName,
		Vector:         embedding,
		Limit:          uint64(limit),
		Filter:         filter,
		WithPayload:    &pb.WithPayloadSelector{SelectorOptions: &pb.WithPayloadSelector_Enable{Enable: true}},
	})

	if err != nil {
		return nil, err
	}

	chunks := make([]domain.MaterialChunk, len(resp.Result))
	for i, point := range resp.Result {
		matID, _ := uuid.Parse(point.Payload["material_id"].GetStringValue())
		pID, _ := uuid.Parse(point.Payload["pod_id"].GetStringValue())

		chunks[i] = domain.MaterialChunk{
			ID:         point.Id.GetUuid(),
			MaterialID: matID,
			PodID:      pID,
			ChunkIndex: int(point.Payload["chunk_index"].GetIntegerValue()),
			Text:       point.Payload["text"].GetStringValue(),
		}
	}

	return chunks, nil
}

func (c *Client) DeleteByMaterialID(ctx context.Context, materialID uuid.UUID) error {
	_, err := c.pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: c.collectionName,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: &pb.Filter{
					Must: []*pb.Condition{
						{
							ConditionOneOf: &pb.Condition_Field{
								Field: &pb.FieldCondition{
									Key: "material_id",
									Match: &pb.Match{
										MatchValue: &pb.Match_Keyword{Keyword: materialID.String()},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	return err
}

func (c *Client) DeleteByPodID(ctx context.Context, podID uuid.UUID) error {
	_, err := c.pointsClient.Delete(ctx, &pb.DeletePoints{
		CollectionName: c.collectionName,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: &pb.Filter{
					Must: []*pb.Condition{
						{
							ConditionOneOf: &pb.Condition_Field{
								Field: &pb.FieldCondition{
									Key: "pod_id",
									Match: &pb.Match{
										MatchValue: &pb.Match_Keyword{Keyword: podID.String()},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	return err
}

// HealthCheck checks if Qdrant is accessible by getting collection info.
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.collectionsClient.Get(ctx, &pb.GetCollectionInfoRequest{
		CollectionName: c.collectionName,
	})
	if err != nil {
		return fmt.Errorf("Qdrant health check failed: %w", err)
	}
	return nil
}
