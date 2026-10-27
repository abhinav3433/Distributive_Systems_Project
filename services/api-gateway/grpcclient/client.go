package grpcclient

import (
	"context"
	"log"
	"net/http"
	"time"

	"api-gateway/config"
	authv1 "proto/auth/v1"
	cartv1 "proto/cart/v1"
	orderv1 "proto/order/v1"
	paymentv1 "proto/payment/v1"
	productv1 "proto/product/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Clients struct {
	AuthClient    authv1.AuthServiceClient
	ProductClient productv1.ProductServiceClient
	CartClient    cartv1.CartServiceClient
	OrderClient   orderv1.OrderServiceClient
	PaymentClient paymentv1.PaymentServiceClient

	authConn    *grpc.ClientConn
	productConn *grpc.ClientConn
	cartConn    *grpc.ClientConn
	orderConn   *grpc.ClientConn
	paymentConn *grpc.ClientConn

	cfg *config.Config
}

func NewClients(cfg *config.Config) (*Clients, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	authConn, err := grpc.Dial(cfg.AuthGRPCAddr, dialOpts...)
	if err != nil {
		log.Printf("[GATEWAY-GRPC] Warning: Failed to dial Auth Service at %s: %v", cfg.AuthGRPCAddr, err)
	}

	productConn, err := grpc.Dial(cfg.ProductGRPCAddr, dialOpts...)
	if err != nil {
		log.Printf("[GATEWAY-GRPC] Warning: Failed to dial Product Service at %s: %v", cfg.ProductGRPCAddr, err)
	}

	cartConn, err := grpc.Dial(cfg.CartGRPCAddr, dialOpts...)
	if err != nil {
		log.Printf("[GATEWAY-GRPC] Warning: Failed to dial Cart Service at %s: %v", cfg.CartGRPCAddr, err)
	}

	orderConn, err := grpc.Dial(cfg.OrderGRPCAddr, dialOpts...)
	if err != nil {
		log.Printf("[GATEWAY-GRPC] Warning: Failed to dial Order Service at %s: %v", cfg.OrderGRPCAddr, err)
	}

	paymentConn, err := grpc.Dial(cfg.PaymentGRPCAddr, dialOpts...)
	if err != nil {
		log.Printf("[GATEWAY-GRPC] Warning: Failed to dial Payment Service at %s: %v", cfg.PaymentGRPCAddr, err)
	}

	return &Clients{
		AuthClient:    authv1.NewAuthServiceClient(authConn),
		ProductClient: productv1.NewProductServiceClient(productConn),
		CartClient:    cartv1.NewCartServiceClient(cartConn),
		OrderClient:   orderv1.NewOrderServiceClient(orderConn),
		PaymentClient: paymentv1.NewPaymentServiceClient(paymentConn),
		authConn:      authConn,
		productConn:   productConn,
		cartConn:      cartConn,
		orderConn:     orderConn,
		paymentConn:   paymentConn,
		cfg:           cfg,
	}, nil
}

func (c *Clients) Close() {
	if c.authConn != nil {
		c.authConn.Close()
	}
	if c.productConn != nil {
		c.productConn.Close()
	}
	if c.cartConn != nil {
		c.cartConn.Close()
	}
	if c.orderConn != nil {
		c.orderConn.Close()
	}
	if c.paymentConn != nil {
		c.paymentConn.Close()
	}
}

// CheckHealth queries the downstream gRPC service standard health check
func CheckServiceHealth(conn *grpc.ClientConn, serviceName string) string {
	if conn == nil {
		return "DOWN"
	}
	healthClient := grpc_health_v1.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: serviceName})
	if err != nil || resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return "DOWN"
	}
	return "UP"
}

func (c *Clients) GetHealthStatus() gin.H {
	return gin.H{
		"auth_grpc":    CheckServiceHealth(c.authConn, "auth.v1.AuthService"),
		"product_grpc": CheckServiceHealth(c.productConn, "product.v1.ProductService"),
		"cart_grpc":    CheckServiceHealth(c.cartConn, "cart.v1.CartService"),
		"order_grpc":   CheckServiceHealth(c.orderConn, "order.v1.OrderService"),
		"payment_grpc": CheckServiceHealth(c.paymentConn, "payment.v1.PaymentService"),
	}
}

// HandleGRPCError maps gRPC status errors to clean JSON responses with standard HTTP status codes
func HandleGRPCError(c *gin.Context, err error, defaultMsg string) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMsg + ": " + err.Error()})
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	case codes.DeadlineExceeded:
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timed out while contacting downstream service"})
	case codes.Unavailable:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Downstream service is currently unavailable", "details": st.Message()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}
