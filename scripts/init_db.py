from datetime import datetime
from email.policy import HTTP
from http import HTTPStatus
import json
from typing import List

import requests
from pydantic import BaseModel, Field, parse_file_as


class RawAppt(BaseModel):
    id: int
    start_time: datetime = Field(alias="started_at")
    end_time: datetime = Field(alias="ended_at")
    user_id: int
    trainer_id: int


def initialize():
    raw_appts = parse_file_as(List[RawAppt], "./data/appointments.json")

    user_ids = {raw_appt.user_id for raw_appt in raw_appts}
    trainer_ids = {raw_appt.trainer_id for raw_appt in raw_appts}

    user_url = "http://localhost:8080/users"
    trainer_url = "http://localhost:8080/trainers"

    for user_id in user_ids:
        r = requests.post(
            user_url,
            json={
                "id": user_id,
                "name": f"User {user_id}",
                "email": f"user_{user_id}@email.com",
            },
        )

        if r.status_code not in [HTTPStatus.CREATED, HTTPStatus.OK]:
            print(f"Failed to create user {user_id}")
            print(r.text)
        else:
            print(f"Created user {user_id}")

    for trainer_id in trainer_ids:
        r = requests.post(
            trainer_url,
            json={
                "id": trainer_id,
                "name": f"Trainer {trainer_id}",
                "email": f"trainer_{trainer_id}@email.com",
            },
        )

        if r.status_code not in [HTTPStatus.CREATED, HTTPStatus.OK]:
            print(f"Failed to create trainer {trainer_id}")
            print(r.text)
        else:
            print(f"Created trainer {trainer_id}")

    appt_url = "http://localhost:8080/appointments"

    for raw_appt in raw_appts:
        r = requests.post(appt_url, data=raw_appt.json())
        if r.status_code not in [HTTPStatus.CREATED, HTTPStatus.OK]:
            print(f"Failed to create appointment {raw_appt.id}")
            print(r.status_code, r.text)
        else:
            print(f"Created appointment {raw_appt.id}")


if __name__ == "__main__":
    initialize()
