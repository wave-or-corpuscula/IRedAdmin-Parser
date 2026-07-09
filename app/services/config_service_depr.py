import json
from typing import List

from app.utils import ServerConfig


class SaveConfigError(Exception):
    pass


class ConfigDosentExistsError(Exception):
    pass


class EmptyConfigFileError(Exception):
    pass


class InvalidConfigDataError(Exception):
    pass


class ConfigService:
    def __init__(self, config_path: str) -> None:
        self.config_path = config_path

    def save_config(self, config: ServerConfig):
        configs = self.load_configs()
        for i in range(len(configs)):
            if configs[i].id == config.id:
                configs[i] = config
                break
        else:
            configs.append(config)

        self.save_configs(configs)

    def save_configs(self, configs: List[ServerConfig]):
        try:
            with open(self.config_path, "w", encoding="utf-8") as f:
                data = [c.to_dict() for c in configs]
                json.dump(data, f, indent=4, ensure_ascii=False)
                return True
        except IOError as e:
            raise SaveConfigError(f"Error while saving config: {e}")

    def load_configs(self) -> List[ServerConfig]:
        try:
            with open(self.config_path, "r") as f:
                data = json.load(f)
        except FileNotFoundError:
            print("Config dosent exists")
            return []
        except json.JSONDecodeError:  # empty config file
            print("Empty config file")
            return []

        configs = []
        try:
            for raw in data:
                configs.append(ServerConfig.from_dict(raw))
        except Exception:
            raise InvalidConfigDataError
        return configs

    def delete_config(self, config: ServerConfig):
        configs = self.load_configs()

        rem_index = -1
        for i in range(len(configs)):
            if configs[i].id == config.id:
                rem_index = i
                break

        if rem_index != -1:
            configs.pop(rem_index)

        self.save_configs(configs)


cfg_service = ConfigService(".iredcreds.json")
