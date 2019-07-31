package pschecker

type cache map[int]cacheItem
type cacheItem struct {
	t []Target
	p Proc
}

// type nocache map[Target]struct{}

type cacheCount map[*Target]int

func (c *cache) in() bool {
	return true

}
func (c cacheCount) in() bool {
	return true
}

// 新しくできる
// → whitelistに含まれるか
// 	含まれる → whitecacheに追加・通知
// 	含まれない → そのまま
// → blacklistに含まれるか
// 	含まれる → blackcacheに追加・通知
// 	含まない → そのまま

// 削除される
// → whitecacheに含まれるか
// 	含まれる → whitecacheから削除・通知
// 		まだ該当するwhitelistを満たす要素はまだあるか
// 			ない → 通知
// 			ある → そのまま
// 	含まれない → そのまま
// → blackcacheに含まれるか
// 	含まれる → blackcacheから削除・通知
// 		該当するblacklistを満たす要素はまだあるか
// 			ない → 通知
// 			ある → そのまま
// 	含まれない → そのまま

// 変更される
// → whitecacheに含まれるか
// 	含まれる → まだ該当するwhitelistに合致するか
// 		合致する → 更新
// 		合致しない → 削除・通知
// 			まだ該当するwhitelistに合致するプロセスが存在するか
// 				存在しない → 通知
// 				存在する → そのまま
// 	含まれない → 該当するno whitelist cacheが存在するか
// 		存在する → 追加・通知
// 		存在しない → そのまま
// → blackcacheに含まれるか
// 	含まれる → まだ該当するblacklistに合致するか
// 		合致する → 更新
// 		合致しない → 削除・通知
// 			まだ該当のblacklistに合致するプロセスが存在するか
// 				存在しない → 通知
// 				存在する → そのまま
// 	含まれない → 該当する no blacklist cacheが存在するか
// 		存在する → 追加・通知
// 		存在しない → そのまま
