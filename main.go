package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	txtPath string
	zipPath string
)

func init() {
	txtPath = "C:\\fyne-packages"
	zipPath = "C:\\fyne-packages"
}

func main() {
	// 디렉토리가 없으면 생성
	os.MkdirAll(txtPath, os.ModePerm)
	os.MkdirAll(zipPath, os.ModePerm)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n깃허브 URL을 입력하세요 (종료하려면 'exit' 입력): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("입력 읽기 오류:", err)
			continue
		}

		// 입력 문자열 정리 (Windows의 경우 \r\n 처리)
		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("프로그램을 종료합니다.")
			break
		}

		// 쉼표로 구분된 URL들을 분리
		urls := strings.Split(input, ",")
		for _, url := range urls {
			// 각 URL 앞뒤 공백 제거
			url = strings.TrimSpace(url)

			// GitHub URL 검증 및 정규화
			if !strings.HasPrefix(url, "https://github.com/") && !strings.HasPrefix(url, "github.com/") {
				fmt.Printf("잘못된 GitHub URL입니다: %s\n", url)
				continue
			}

			// URL에서 프로젝트 이름 추출
			parts := strings.Split(url, "/")
			startIdx := 3
			if strings.HasPrefix(url, "github.com/") {
				startIdx = 1
			}
			if len(parts) < startIdx+2 {
				fmt.Println("잘못된 URL 형식입니다.")
				continue
			}

			projectName := fmt.Sprintf("%s_%s", parts[startIdx], parts[startIdx+1])
			fmt.Printf("의존성 제목: go/%s/%s\n", parts[startIdx], parts[startIdx+1]) // 추가된 부분

			// git clone을 위한 URL 생성
			gitURL := url
			if !strings.HasPrefix(gitURL, "https://") {
				gitURL = "https://" + gitURL
			}
			if !strings.HasSuffix(gitURL, ".git") {
				gitURL = gitURL + ".git"
			}

			// 파일에 쓸 URL 생성 (input 대신 사용)
			fileURL := url
			if !strings.HasPrefix(fileURL, "https://") {
				fileURL = "https://" + fileURL
			}

			// path.txt 파일 경로 설정
			pathTxtFile := filepath.Join(txtPath, "path.txt")
			fmt.Printf("path.txt 파일 업데이트 중... (%s)\n", pathTxtFile)

			// 파일이 없으면 생성, 있으면 추가 모드로 열기
			f, err := os.OpenFile(pathTxtFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("파일 열기 오류:", err)
				continue
			}

			// 새 줄 추가
			if _, err := f.WriteString("\r\n\r\n" + projectName + "\r\n" + fileURL + "\r\n"); err != nil {
				fmt.Println("파일 쓰기 오류:", err)
				f.Close()
				continue
			}
			f.Close()
			fmt.Println("path.txt 파일 업데이트 완료")

			// git clone 경로 설정
			clonePath := filepath.Join(zipPath, projectName)

			// 디렉토리가 이미 존재하는지 확인
			if _, err := os.Stat(clonePath); !os.IsNotExist(err) {
				fmt.Printf("경로에 이미 '%s' 디렉토리가 존재합니다.\n", projectName)
				continue
			}

			fmt.Printf("Git repository 클론 중... (%s)\n", clonePath)

			// Git clone with all branches
			// cmd := exec.Command("git", "clone", "--recursive", gitURL, clonePath)
			// if err := cmd.Run(); err != nil {
			// 	fmt.Println("Git clone 오류:", err)
			// 	continue
			// }

			cmd := exec.Command("git", "clone", "--recursive", gitURL, clonePath)
			msg, err := cmd.CombinedOutput() // Output()대신 CombinedOutput() 사용
			if err != nil {
				fmt.Printf("Git clone 오류: %v\n에러 메시지: %s\n", err, string(msg))
				continue
			}

			// Fetch all branches
			cmd = exec.Command("git", "-C", clonePath, "remote", "update", "--prune")
			if err := cmd.Run(); err != nil {
				fmt.Println("Git fetch 오류:", err)
				continue
			}

			// Checkout all remote branches locally
			cmd = exec.Command("git", "-C", clonePath, "branch", "-r")
			output, err := cmd.Output()
			if err != nil {
				fmt.Println("브랜치 목록 조회 오류:", err)
				continue
			}

			// Create local branches
			branches := strings.Split(string(output), "\n")
			for _, branch := range branches {
				branch = strings.TrimSpace(branch)
				if branch == "" || strings.Contains(branch, "->") || strings.Contains(branch, "HEAD") {
					continue
				}

				localBranch := strings.TrimPrefix(branch, "origin/")
				checkoutCmd := exec.Command("git", "-C", clonePath, "checkout", "-b", localBranch, branch)
				if err := checkoutCmd.Run(); err != nil {
					fmt.Printf("Warning: 브랜치 '%s' 체크아웃 실패: %v\n", localBranch, err)
					continue
				}
			}

			fmt.Println("Git clone 및 모든 브랜치 체크아웃 완료")

			// PowerShell을 사용하여 ZIP 파일 생성
			fmt.Println("ZIP 파일 생성 중...")
			zipPath := clonePath + ".zip"

			// PowerShell Compress-Archive 명령어 사용
			powershellCmd := fmt.Sprintf("Get-ChildItem -Path '%s' -Force -Recurse | Compress-Archive -DestinationPath '%s' -Force",
				clonePath, zipPath)
			cmd = exec.Command("powershell", "-Command", powershellCmd)

			if err := cmd.Run(); err != nil {
				fmt.Println("ZIP 파일 생성 오류:", err)
				continue
			}
			fmt.Printf("ZIP 파일 생성 완료: %s\n", zipPath)
		}

	}
}
