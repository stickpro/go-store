package constant

type ProductReviewStatus string // @name ProductReviewStatus

const (
	ReviewPending  ProductReviewStatus = "PENDING"
	ReviewApproved ProductReviewStatus = "APPROVED"
	ReviewRejected ProductReviewStatus = "REJECTED"
)

func (s ProductReviewStatus) String() string {
	return string(s)
}
