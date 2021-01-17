package app

import (
	"sort"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	Game     *Game  `json:"-"`
	Name     string `json:"name"`
	IsAuthor bool   `json:"is_author"`

	ws       *websocket.Conn
	send     chan *wireMessage
	gameplay *gameplay
}

func (p *Player) closeWithDelay() {
	time.Sleep(wireWriteTimeout)
	close(p.send)
}

type broadcastMessage struct {
	Game    *Game
	Message *wireMessage
}

type Pool struct {
	players    map[*Player]struct{}
	register   chan *Player
	unregister chan *Player
	broadcast  chan *broadcastMessage
}

func (p *Pool) getPlayers(game *Game) []*Player {
	players := make([]*Player, 0)
	for player := range p.players {
		if player.Game.ID == game.ID {
			players = append(players, player)
		}
	}
	return players
}

func (p *Pool) Run() {
	for {
		select {
		case player := <-p.register:
			var gp *gameplay

			players := p.getPlayers(player.Game)
			for _, _player := range players {
				if gp == nil && _player.gameplay != nil && _player.gameplay.state != gpsFinished {
					gp = _player.gameplay
				}
				if _player.Name == player.Name {
					player.send <- &wireMessage{Type: wmtPlayerExists}
					go player.closeWithDelay()
					goto _continue
				}
			}

			if gp == nil {
				if player.IsAuthor {
					gp = newGameplay(player.Game)
				} else {
					player.send <- &wireMessage{Type: wmtNotReady}
					go player.closeWithDelay()
					goto _continue
				}
			}

			for _, _player := range players {
				_player.send <- &wireMessage{
					Type: wmtPlayerRegistered,
					Data: player,
				}
			}

			players = append(players, player)
			sort.Slice(players, func(i, j int) bool {
				return players[i].Name < players[j].Name
			})
			player.send <- &wireMessage{
				Type: wmtReady,
				Data: map[string]interface{}{
					"name":         player.Name,
					"players":      players,
					"gp_state":     gp.state,
					"gp_num_tasks": len(gp.tasks),
				},
			}

			player.gameplay = gp
			player.gameplay.Init(player)

			p.players[player] = struct{}{}
		case player := <-p.unregister:
			if _, ok := p.players[player]; ok {
				delete(p.players, player)
				close(player.send)

				delete(player.gameplay.scores, player)

				players := p.getPlayers(player.Game)
				for _, _player := range players {
					_player.send <- &wireMessage{
						Type: wmtPlayerUnregistered,
						Data: player,
					}
				}
			}
		case bm := <-p.broadcast:
			players := p.getPlayers(bm.Game)
			for _, player := range players {
				select {
				case player.send <- bm.Message:
				default:
					delete(p.players, player)
					close(player.send)
				}
			}
		}
	_continue:
	}
}

func NewPool() *Pool {
	return &Pool{
		players:    make(map[*Player]struct{}),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		broadcast:  make(chan *broadcastMessage),
	}
}
