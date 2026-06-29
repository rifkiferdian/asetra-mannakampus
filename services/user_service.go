package services

import (
	"database/sql"
	"errors"
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo *repositories.UserRepository
}

func (s *UserService) GetUsers() ([]models.User, error) {
	return s.Repo.GetAll()
}

func (s *UserService) GetUserProfile(id int) (models.User, error) {
	if id <= 0 {
		return models.User{}, errors.New("user tidak valid")
	}
	return s.Repo.GetByID(id)
}

// CreateUser memproses data dari form, melakukan validasi dasar, hashing password,
// lalu menyimpan user beserta role yang dipilih.
func (s *UserService) CreateUser(input models.UserCreateInput) error {
	username := strings.TrimSpace(input.Username)
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	status := strings.TrimSpace(input.Status)

	if username == "" || name == "" || input.Password == "" {
		return errors.New("nama, username, dan password wajib diisi")
	}
	if email == "" {
		return errors.New("email wajib diisi")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("email tidak valid")
	}
	if input.NIP <= 0 {
		return errors.New("NIP wajib diisi")
	}
	if status != "active" && status != "non_active" {
		status = "active"
	}

	exists, err := s.Repo.ExistsByUsername(username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("username '%s' sudah digunakan", username)
	}

	exists, err = s.Repo.ExistsByNIP(input.NIP)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("NIP %d sudah digunakan", input.NIP)
	}

	if email != "" {
		exists, err = s.Repo.ExistsByEmail(email)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("email %s sudah digunakan", email)
		}
	}

	roleNames := uniqueStrings(input.RoleNames)
	roleMap, err := s.Repo.GetRoleIDsByNames(roleNames)
	if err != nil {
		return err
	}

	var (
		roleIDs      []int64
		missingRoles []string
	)

	for _, roleName := range roleNames {
		if id, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, id)
		} else {
			missingRoles = append(missingRoles, roleName)
		}
	}

	if len(missingRoles) > 0 {
		return fmt.Errorf("role tidak ditemukan: %s", strings.Join(missingRoles, ", "))
	}

	storeIDs := uniqueInts(input.StoreIDs)
	if storeIDs == nil {
		storeIDs = []int{}
	}
	if len(storeIDs) == 0 {
		return errors.New("store wajib dipilih")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.Repo.CreateUserWithRoles(repositories.UserCreateParams{
		NIP:            input.NIP,
		Username:       username,
		HashedPassword: string(hashedPassword),
		Name:           name,
		Email:          email,
		Status:         status,
		StoreIDs:       storeIDs,
	}, roleIDs)

	return err
}

// UpdateUser memperbarui data user yang sudah ada.
func (s *UserService) UpdateUser(input models.UserUpdateInput) error {
	username := strings.TrimSpace(input.Username)
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	status := strings.TrimSpace(input.Status)

	if input.ID <= 0 {
		return errors.New("user tidak valid")
	}
	if username == "" || name == "" {
		return errors.New("nama dan username wajib diisi")
	}
	if email == "" {
		return errors.New("email wajib diisi")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("email tidak valid")
	}
	if input.NIP <= 0 {
		return errors.New("NIP wajib diisi")
	}
	if status != "active" && status != "non_active" {
		status = "active"
	}

	exists, err := s.Repo.ExistsByUsernameExceptID(username, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("username '%s' sudah digunakan", username)
	}

	exists, err = s.Repo.ExistsByNIPExceptID(input.NIP, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("NIP %d sudah digunakan", input.NIP)
	}

	if email != "" {
		exists, err = s.Repo.ExistsByEmailExceptID(email, input.ID)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("email %s sudah digunakan", email)
		}
	}

	roleNames := uniqueStrings(input.RoleNames)
	roleMap, err := s.Repo.GetRoleIDsByNames(roleNames)
	if err != nil {
		return err
	}

	var (
		roleIDs      []int64
		missingRoles []string
	)

	for _, roleName := range roleNames {
		if id, ok := roleMap[roleName]; ok {
			roleIDs = append(roleIDs, id)
		} else {
			missingRoles = append(missingRoles, roleName)
		}
	}

	if len(missingRoles) > 0 {
		return fmt.Errorf("role tidak ditemukan: %s", strings.Join(missingRoles, ", "))
	}

	storeIDs := uniqueInts(input.StoreIDs)
	if storeIDs == nil {
		storeIDs = []int{}
	}
	if len(storeIDs) == 0 {
		return errors.New("store wajib dipilih")
	}

	var hashedPassword string
	if strings.TrimSpace(input.Password) != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hashedPassword = string(hashed)
	}

	return s.Repo.UpdateUserWithRoles(repositories.UserUpdateParams{
		ID:             input.ID,
		NIP:            input.NIP,
		Username:       username,
		HashedPassword: hashedPassword,
		Name:           name,
		Email:          email,
		Status:         status,
		StoreIDs:       storeIDs,
	}, roleIDs)
}

// UpdateOwnProfile memperbarui data diri user login tanpa mengubah role, store, atau status.
func (s *UserService) UpdateOwnProfile(input models.UserProfileUpdateInput) error {
	username := strings.TrimSpace(input.Username)
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)

	if input.ID <= 0 {
		return errors.New("user tidak valid")
	}
	if username == "" || name == "" {
		return errors.New("nama dan username wajib diisi")
	}
	if email == "" {
		return errors.New("email wajib diisi")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("email tidak valid")
	}

	exists, err := s.Repo.ExistsByUsernameExceptID(username, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("username '%s' sudah digunakan", username)
	}

	exists, err = s.Repo.ExistsByEmailExceptID(email, input.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email %s sudah digunakan", email)
	}

	return s.Repo.UpdateOwnProfile(input.ID, username, name, email)
}

// UpdateOwnPassword mengganti password user login setelah password lama diverifikasi.
func (s *UserService) UpdateOwnPassword(input models.UserPasswordUpdateInput) error {
	if input.ID <= 0 {
		return errors.New("user tidak valid")
	}
	if strings.TrimSpace(input.CurrentPassword) == "" || strings.TrimSpace(input.NewPassword) == "" {
		return errors.New("password lama dan password baru wajib diisi")
	}
	if input.NewPassword != input.ConfirmPassword {
		return errors.New("konfirmasi password baru tidak sesuai")
	}
	if len(input.NewPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	if input.CurrentPassword == input.NewPassword {
		return errors.New("password baru harus berbeda dari password lama")
	}

	currentHash, err := s.Repo.GetPasswordHashByID(input.ID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(input.CurrentPassword)) != nil {
		return errors.New("password lama tidak sesuai")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.Repo.UpdatePassword(input.ID, string(hashed))
}

const userModelType = "Models\\User"

// DeleteUser removes user data by ID.
func (s *UserService) DeleteUser(id int) error {
	if id <= 0 {
		return errors.New("user id tidak valid")
	}
	return s.Repo.DeleteUser(id)
}

func UserHasPermission(userID int, perm string) (bool, error) {
	var dummy int
	// Cek permission via role yang dimiliki user
	queryRole := `
		SELECT 1
		FROM model_has_roles mhr
		JOIN role_has_permissions rhp ON rhp.role_id = mhr.role_id
		JOIN permissions p ON p.id = rhp.permission_id
		WHERE mhr.model_id = ? AND mhr.model_type = ? AND p.name = ?
		LIMIT 1
	`
	err := config.DB.QueryRow(queryRole, userID, userModelType, perm).Scan(&dummy)
	if err == nil {
		return true, nil
	}
	if err != sql.ErrNoRows {
		return false, err
	}

	// Fallback: cek permission langsung ke user (model_has_permissions)
	queryDirect := `
		SELECT 1
		FROM model_has_permissions mhp
		JOIN permissions p ON p.id = mhp.permission_id
		WHERE mhp.model_id = ? AND mhp.model_type = ? AND p.name = ?
		LIMIT 1
	`

	err = config.DB.QueryRow(queryDirect, userID, userModelType, perm).Scan(&dummy)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil // tidak punya permission
	}

	return false, err // error lain
}

func GetUserPermissions(userID int) (map[string]bool, error) {
	perms := make(map[string]bool)

	rows, err := config.DB.Query(`
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_has_permissions rhp ON rhp.permission_id = p.id
		JOIN model_has_roles mhr ON mhr.role_id = rhp.role_id
		WHERE mhr.model_id = ? AND mhr.model_type = ?

		UNION

		SELECT DISTINCT p2.name
		FROM permissions p2
		JOIN model_has_permissions mhp ON mhp.permission_id = p2.id
		WHERE mhp.model_id = ? AND mhp.model_type = ?
	`, userID, userModelType, userID, userModelType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms[name] = true
	}

	return perms, nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		result = append(result, v)
	}
	return result
}

func uniqueInts(values []int) []int {
	seen := make(map[int]bool)
	var result []int
	for _, v := range values {
		if seen[v] {
			continue
		}
		seen[v] = true
		result = append(result, v)
	}
	return result
}
