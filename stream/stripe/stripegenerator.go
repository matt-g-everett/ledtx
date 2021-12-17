package stripe

type StripeGenerator interface {
	CreateStripe() Stripe
}
