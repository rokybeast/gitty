package git

import (
	"os/exec"
	"strings"
)

type CommitNode struct {
	Hash    string
	Parents []string
	Message string
}

type GraphRow struct {
	Commit    CommitNode
	Prefix    string // The graph structure string
	NodeGlyph string // The glyph for the node (set by UI)
	IsRoute   bool   // True if this is just a routing line
}

func GetCommits() ([]CommitNode, error) {
	cmd := exec.Command("git", "log", "--all", "--topo-order", "--format=%H%x00%P%x00%s")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []CommitNode
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) != 3 {
			continue
		}
		hash := parts[0]
		parentsStr := strings.TrimSpace(parts[1])
		var parents []string
		if parentsStr != "" {
			parents = strings.Split(parentsStr, " ")
		}
		msg := parts[2]
		commits = append(commits, CommitNode{
			Hash:    hash,
			Parents: parents,
			Message: msg,
		})
	}
	return commits, nil
}

func BuildGraph() ([]GraphRow, error) {
	commits, err := GetCommits()
	if err != nil {
		return nil, err
	}

	var rows []GraphRow
	var activeTracks []string

	for _, commit := range commits {
		// 1. Find column(s) for the current commit
		var cols []int
		for j, t := range activeTracks {
			if t == commit.Hash {
				cols = append(cols, j)
			}
		}

		col := -1
		if len(cols) > 0 {
			col = cols[0]
			// Merge route if needed
			if len(cols) > 1 {
				var mergeBuilder strings.Builder
				maxCol := cols[len(cols)-1]
				for j := 0; j <= maxCol; j++ {
					if j < col {
						if activeTracks[j] != "" {
							mergeBuilder.WriteString("│ ")
						} else {
							mergeBuilder.WriteString("  ")
						}
					} else if j == col {
						mergeBuilder.WriteString("├─")
					} else if j < maxCol {
						if activeTracks[j] != "" && activeTracks[j] != commit.Hash {
							mergeBuilder.WriteString("┼─")
						} else {
							mergeBuilder.WriteString("──")
						}
					} else if j == maxCol {
						mergeBuilder.WriteString("╯ ")
					}
				}
				rows = append(rows, GraphRow{
					Prefix:  mergeBuilder.String(),
					IsRoute: true,
				})

				for i := 1; i < len(cols); i++ {
					activeTracks[cols[i]] = ""
				}
				for len(activeTracks) > 0 && activeTracks[len(activeTracks)-1] == "" {
					activeTracks = activeTracks[:len(activeTracks)-1]
				}
			}
		} else {
			// New track
			for j := 0; j < len(activeTracks); j++ {
				if activeTracks[j] == "" {
					col = j
					break
				}
			}
			if col == -1 {
				col = len(activeTracks)
				activeTracks = append(activeTracks, commit.Hash)
			} else {
				activeTracks[col] = commit.Hash
			}
		}

		// 2. Commit row
		var graph strings.Builder
		for j := 0; j < len(activeTracks); j++ {
			if j == col {
				graph.WriteString("* ") 
			} else if activeTracks[j] != "" {
				graph.WriteString("│ ")
			} else {
				graph.WriteString("  ")
			}
		}
		rows = append(rows, GraphRow{
			Commit:  commit,
			Prefix:  graph.String(),
			IsRoute: false,
		})

		// 3. Update active tracks for parents
		numParents := len(commit.Parents)
		if numParents == 0 {
			activeTracks[col] = ""
		} else if numParents == 1 {
			activeTracks[col] = commit.Parents[0]
		} else if numParents >= 2 {
			activeTracks[col] = commit.Parents[0]
			
			p2col := -1
			for j := col + 1; j < len(activeTracks); j++ {
				if activeTracks[j] == "" {
					p2col = j
					break
				}
			}
			if p2col == -1 {
				p2col = len(activeTracks)
				activeTracks = append(activeTracks, commit.Parents[1])
			} else {
				activeTracks[p2col] = commit.Parents[1]
			}

			var routeBuilder strings.Builder
			for j := 0; j <= p2col; j++ {
				if j < col {
					if activeTracks[j] != "" {
						routeBuilder.WriteString("│ ")
					} else {
						routeBuilder.WriteString("  ")
					}
				} else if j == col {
					routeBuilder.WriteString("├─")
				} else if j < p2col {
					if activeTracks[j] != "" {
						routeBuilder.WriteString("┼─")
					} else {
						routeBuilder.WriteString("──")
					}
				} else if j == p2col {
					routeBuilder.WriteString("╮ ")
				}
			}
			rows = append(rows, GraphRow{
				Prefix:  routeBuilder.String(),
				IsRoute: true,
			})
		}
		
		for len(activeTracks) > 0 && activeTracks[len(activeTracks)-1] == "" {
			activeTracks = activeTracks[:len(activeTracks)-1]
		}
	}

	return rows, nil
}
