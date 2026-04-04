package gamemap

import (
	"container/list"
	"fmt"
)

type PathNode struct {
	X, Y    int
	G, H, F int
	Parent  *PathNode
}

type PathFinder struct {
	MaxSteps  int
	AllowDiag bool
}

func NewPathFinder() *PathFinder {
	return &PathFinder{
		MaxSteps:  2000,
		AllowDiag: true,
	}
}

func (pf *PathFinder) FindPath(gameMap *GameMap, startX, startY, endX, endY int) []*PathNode {
	if gameMap == nil {
		return nil
	}

	if !gameMap.CanWalk(startX, startY) || !gameMap.CanWalk(endX, endY) {
		return nil
	}

	if startX == endX && startY == endY {
		return []*PathNode{{X: startX, Y: startY}}
	}

	openList := list.New()
	closedMap := make(map[string]*PathNode)

	startNode := &PathNode{
		X: startX,
		Y: startY,
		G: 0,
		H: pf.heuristic(startX, startY, endX, endY),
	}
	startNode.F = startNode.G + startNode.H
	openList.PushBack(startNode)
	closedMap[pf.key(startX, startY)] = startNode

	steps := 0
	for openList.Len() > 0 && steps < pf.MaxSteps {
		steps++

		currentElem := pf.getLowestF(openList)
		current := currentElem.Value.(*PathNode)
		openList.Remove(currentElem)

		if current.X == endX && current.Y == endY {
			return pf.reconstructPath(current)
		}

		neighbors := pf.getNeighbors(gameMap, current)
		for _, neighbor := range neighbors {
			neighborKey := pf.key(neighbor.X, neighbor.Y)

			if _, exists := closedMap[neighborKey]; exists {
				continue
			}

			tentativeG := current.G + 1

			inOpen := pf.isInOpenList(openList, neighbor.X, neighbor.Y)
			if !inOpen || tentativeG < neighbor.G {
				neighbor.G = tentativeG
				neighbor.H = pf.heuristic(neighbor.X, neighbor.Y, endX, endY)
				neighbor.F = neighbor.G + neighbor.H
				neighbor.Parent = current

				if !inOpen {
					openList.PushBack(neighbor)
				}
			}
		}

		currentKey := pf.key(current.X, current.Y)
		closedMap[currentKey] = current
	}

	return nil
}

func (pf *PathFinder) key(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

func (pf *PathFinder) heuristic(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (pf *PathFinder) getLowestF(l *list.List) *list.Element {
	var lowestElem *list.Element
	for e := l.Front(); e != nil; e = e.Next() {
		node := e.Value.(*PathNode)
		if lowestElem == nil {
			lowestElem = e
		} else {
			lowestNode := lowestElem.Value.(*PathNode)
			if node.F < lowestNode.F {
				lowestElem = e
			}
		}
	}
	return lowestElem
}

func (pf *PathFinder) getNeighbors(gameMap *GameMap, node *PathNode) []*PathNode {
	var neighbors []*PathNode

	dirs := [][2]int{
		{0, -1}, {1, -1}, {1, 0}, {1, 1},
		{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
	}

	for _, dir := range dirs {
		nx := node.X + dir[0]
		ny := node.Y + dir[1]

		if gameMap.CanWalk(nx, ny) {
			neighbors = append(neighbors, &PathNode{X: nx, Y: ny})
		}
	}

	return neighbors
}

func (pf *PathFinder) isInOpenList(l *list.List, x, y int) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		node := e.Value.(*PathNode)
		if node.X == x && node.Y == y {
			return true
		}
	}
	return false
}

func (pf *PathFinder) reconstructPath(node *PathNode) []*PathNode {
	var path []*PathNode
	for node != nil {
		path = append([]*PathNode{node}, path...)
		node = node.Parent
	}
	return path
}
