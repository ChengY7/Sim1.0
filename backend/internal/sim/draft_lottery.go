package sim

import "math/rand"

type DraftLotteryResult struct {
	Picks []DraftPick `json:"picks"`
}

type DraftPick struct {
	Pick   int    `json:"pick"`
	TeamID string `json:"team_id"`
	Seed   int    `json:"seed"`
}

// SimulateDraftLottery runs the NBA draft lottery for 14 teams.
// teams[0] = seed 1 (worst record), teams[13] = seed 14.
// combinations[i] is the number of lottery ball combinations for seed i+1.
// Picks 1-4 are determined by weighted lottery; picks 5-14 go to remaining
// teams in seed order (worst record first).
func SimulateDraftLottery(teams [14]string, combinations [14]int, seed int64) DraftLotteryResult {
	rng := rand.New(rand.NewSource(seed))

	remaining := combinations
	assigned := [14]bool{}
	lotteryWinners := make([]int, 0, 4)

	for range 4 {
		total := 0
		for _, c := range remaining {
			total += c
		}

		r := rng.Intn(total)
		cumulative := 0
		winner := -1
		for i, c := range remaining {
			cumulative += c
			if r < cumulative {
				winner = i
				break
			}
		}

		lotteryWinners = append(lotteryWinners, winner)
		assigned[winner] = true
		remaining[winner] = 0
	}

	picks := make([]DraftPick, 14)

	for pickIdx, seedIdx := range lotteryWinners {
		picks[pickIdx] = DraftPick{
			Pick:   pickIdx + 1,
			TeamID: teams[seedIdx],
			Seed:   seedIdx + 1,
		}
	}

	pickNum := 5
	for seedIdx := 0; seedIdx < 14; seedIdx++ {
		if !assigned[seedIdx] {
			picks[pickNum-1] = DraftPick{
				Pick:   pickNum,
				TeamID: teams[seedIdx],
				Seed:   seedIdx + 1,
			}
			pickNum++
		}
	}

	return DraftLotteryResult{Picks: picks}
}
