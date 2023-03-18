CREATE TABLE IF NOT EXISTS challenges (
  id SERIAL NOT NULL,
  user_id VARCHAR(255) NOT NULL UNIQUE,
  challenge VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  FOREIGN KEY (user_id) REFERENCES users (uid)
);

CREATE TRIGGER refresh_challenges_updated_at_step1
  BEFORE UPDATE ON challenges FOR EACH ROW
  EXECUTE PROCEDURE refresh_updated_at_step1();
CREATE TRIGGER refresh_chalenges_updated_at_step2
  BEFORE UPDATE OF updated_at ON challenges FOR EACH ROW
  EXECUTE PROCEDURE refresh_updated_at_step2();
CREATE TRIGGER refresh_challenges_updated_at_step3
  BEFORE UPDATE ON users FOR EACH ROW
  EXECUTE PROCEDURE refresh_updated_at_step3();