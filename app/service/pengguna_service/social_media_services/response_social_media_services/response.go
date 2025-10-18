package response_social_media_pengguna

type ResponseEngageSocialMedia struct {
	Message string `json:"pesan_ubah_social_media"`
}

type ResponseFollowSeller struct {
	Message string `json:"pesan_follow_seller"`
}

type ResponseUnfollowSeller struct {
	Message string `json:"pesan_unfollow_seller"`
}
