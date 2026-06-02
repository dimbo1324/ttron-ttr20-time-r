package emu

import (
	"bytes"
	"log"
	"math/rand"
	"net"
	"runtime/debug"
	"sln/internal/config"
	"sln/internal/emulator"
	"sln/internal/frame"
	"sln/internal/util"
	"time"
)

// handleConnection обслуживает одно TCP-соединение
// Защищён от паники, читает байты, собирает фреймы и отвечает
func handleConnection(conn net.Conn, cfg *config.Config, logger *log.Logger) {
	// recover чтобы паника в обработчике не убивала весь сервер
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("[%s] PANIC recovered: %v\n%s", conn.RemoteAddr(), r, string(debug.Stack()))
			_ = conn.Close()
		}
	}()

	// Закрываем соединение по выходу и логируем окончание хендлера
	defer func() {
		_ = conn.Close()
		logger.Printf("[%s] connection handler finished", conn.RemoteAddr())
	}()

	var buf bytes.Buffer
	tmp := make([]byte, 4096)
	readTimeout := time.Duration(cfg.ReadTimeout) * time.Second

	for {
		// Устанавливаем deadline для чтения
		_ = conn.SetReadDeadline(time.Now().Add(readTimeout))
		n, err := conn.Read(tmp)
		if err != nil {
			// При ошибке чтения закрываем хендлер
			logger.Printf("[%s] read error: %v", conn.RemoteAddr(), err)
			return
		}
		if n == 0 {
			continue
		}
		// Пишем полученные байты в буфер для парсинга фреймов
		buf.Write(tmp[:n])

		// Пока есть полный фрейм - извлекаем и обрабатываем
		for {
			frameBytes, ok := frame.ExtractFrame(&buf)
			if !ok {
				break
			}
			logger.Printf("[%s] RX: %s", conn.RemoteAddr(), util.HexDump(frameBytes))

			// Проверяем контрольную сумму/формат фрейма
			if err := frame.VerifyFrame(frameBytes); err != nil {
				logger.Printf("[%s] frame verification failed: %v", conn.RemoteAddr(), err)
				// Игнорируем некорректный фрейм и ждём следующий
				continue
			}

			// Базовый разбор: control, addr, data (если есть).
			if len(frameBytes) < 6 {
				logger.Printf("[%s] frame too short", conn.RemoteAddr())
				continue
			}
			control := frameBytes[3]
			addr := frameBytes[4]
			data := frame.PayloadData(frameBytes)
			var cmd byte
			if len(data) > 0 {
				cmd = data[0]
			}

			// Обработка известных команд
			switch cmd {
			case 0x01:
				// Команда чтения времени
				logger.Printf("[%s] read-time request (ctrl=0x%02X addr=0x%02X)", conn.RemoteAddr(), control, addr)
				resp := emulator.BuildTimeResponse(control, addr, data, cfg.CRCMode, byte(cfg.AdapterAddr))

				// Опциональня искусственная задержка для тестов
				if cfg.DelayMs > 0 {
					time.Sleep(time.Duration(cfg.DelayMs) * time.Millisecond)
				}

				// Иногда инжектим плохой CRC (для тестирования).
				if rand.Float64() < cfg.BadCRCProb {
					logger.Printf("[%s] injecting bad CRC", conn.RemoteAddr())
					frame.CorruptChecksum(resp, cfg.CRCMode)
				}

				// Иногда фрагментируем ответ на две части
				if rand.Float64() < cfg.FragProb && len(resp) > 1 {
					i := len(resp) / 2
					if i < 1 {
						i = 1
					}
					logger.Printf("[%s] sending fragmented response (%d + %d)", conn.RemoteAddr(), i, len(resp)-i)
					if _, err := conn.Write(resp[:i]); err != nil {
						logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
						return
					}
					time.Sleep(40 * time.Millisecond)
					if _, err := conn.Write(resp[i:]); err != nil {
						logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
						return
					}
				} else {
					// Отправляем полный ответ
					if _, err := conn.Write(resp); err != nil {
						logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
						return
					}
				}
				logger.Printf("[%s] TX: %s", conn.RemoteAddr(), util.HexDump(resp))
			default:
				// Для неизвестных команд отправляем ACK/echo
				logger.Printf("[%s] generic/unknown cmd 0x%02X - sending ACK", conn.RemoteAddr(), cmd)
				resp := emulator.BuildAckResponse(control, addr, data, cfg.CRCMode, byte(cfg.AdapterAddr))
				if cfg.DelayMs > 0 {
					time.Sleep(time.Duration(cfg.DelayMs) * time.Millisecond)
				}
				if rand.Float64() < cfg.BadCRCProb {
					frame.CorruptChecksum(resp, cfg.CRCMode)
				}
				if rand.Float64() < cfg.FragProb && len(resp) > 1 {
					i := len(resp) / 2
					if i < 1 {
						i = 1
					}
					if _, err := conn.Write(resp[:i]); err != nil {
						logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
						return
					}
					time.Sleep(40 * time.Millisecond)
					if _, err := conn.Write(resp[i:]); err != nil {
						logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
						return
					}
				} else {
					if _, err := conn.Write(resp); err != nil {
						logger.Printf("[%s] write error: %v", conn.RemoteAddr(), err)
						return
					}
				}
				logger.Printf("[%s] TX: %s", conn.RemoteAddr(), util.HexDump(resp))
			}
		}
	}
}
