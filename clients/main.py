import json
import re
import threading
import tkinter as tk
from tkinter import messagebox, ttk

from pitaya_client import PitayaClient

HOST = "172.31.11.52"
PORT = 8080
ACCOUNT_ID_PATTERN = re.compile(r"^[A-Za-z0-9]+$")
PASSWORD_PATTERN = re.compile(r"^[A-Za-z0-9@#$%^&*()_+=!?.\-,]{1,16}$")

class GameClientWindow:
	def __init__(self, root: tk.Tk):
		self.root = root
		self.root.title("Card Game Login")
		self.root.resizable(False, False)

		self.client = PitayaClient(f"ws://{HOST}:{PORT}/")
		# self.client.register_route("Summon.Notify", self._handle_server_push)
		self.login_in_progress = False
		self.action_in_progress = False

		self.account_id = tk.StringVar()
		self.password = tk.StringVar()
		self.register_account_id = tk.StringVar()
		self.register_password = tk.StringVar()
		self.uid = None
		self.status = tk.StringVar(value="Disconnected")
		self.game_status = tk.StringVar(value="Ready")
		self.login_frame = None
		self.register_frame = None
		self.game_frame = None

		self._build_ui()
		self._build_register_frame()
		self.register_frame.grid_remove()
		self.root.protocol("WM_DELETE_WINDOW", self.close)

	def _build_ui(self):
		self.login_frame = ttk.Frame(self.root, padding=16)
		self.login_frame.grid(row=0, column=0)

		ttk.Label(self.login_frame, text="Account").grid(
			row=0, column=0, padx=(0, 8), pady=6, sticky="w"
		)
		account_entry = ttk.Entry(self.login_frame, textvariable=self.account_id, width=28)
		account_entry.grid(row=0, column=1, pady=6)

		ttk.Label(self.login_frame, text="Password").grid(
			row=1, column=0, padx=(0, 8), pady=6, sticky="w"
		)
		password_entry = ttk.Entry(
			self.login_frame,
			textvariable=self.password,
			width=28,
			show="*",
		)
		password_entry.grid(row=1, column=1, pady=6)

		self.login_button = ttk.Button(
			self.login_frame,
			text="Login",
			command=self.login,
		)
		self.login_button.grid(row=2, column=0, columnspan=2, pady=(12, 6))

		ttk.Button(
			self.login_frame,
			text="Create account",
			command=self.show_register_frame,
		).grid(row=3, column=0, columnspan=2, pady=6)

		ttk.Label(self.login_frame, textvariable=self.status).grid(
			row=4, column=0, columnspan=2, pady=6
		)

		account_entry.focus_set()
		password_entry.bind("<Return>", lambda _event: self.login())

	def _build_register_frame(self):
		self.register_frame = ttk.Frame(self.root, padding=16)
		self.register_frame.grid(row=0, column=0)

		ttk.Label(self.register_frame, text="Account ID").grid(
			row=0, column=0, padx=(0, 8), pady=6, sticky="w"
		)
		tk.Entry(
			self.register_frame,
			textvariable=self.register_account_id,
			width=28,
		).grid(row=0, column=1, pady=6)

		tk.Label(self.register_frame, text="Password").grid(
			row=1, column=0, padx=(0, 8), pady=6, sticky="w"
		)
		tk.Entry(
			self.register_frame,
			textvariable=self.register_password,
			width=28,
			show="*",
		).grid(row=1, column=1, pady=6)

		self.register_button = ttk.Button(
			self.register_frame,
			text="Create account",
			command=self.create_account,
		)
		self.register_button.grid(row=2, column=0, columnspan=2, pady=(12, 6))

		ttk.Button(
			self.register_frame,
			text="Back to login",
			command=self.show_login_frame,
		).grid(row=3, column=0, columnspan=2, pady=6)

		self.register_status = tk.StringVar(value="")
		ttk.Label(self.register_frame, textvariable=self.register_status).grid(
			row=4, column=0, columnspan=2, pady=6
		)

	def show_register_frame(self):
		self.login_frame.grid_remove()
		self.register_frame.grid()
		self.root.title("Card Game - Create Account")

	def show_login_frame(self):
		self.register_frame.grid_remove()
		self.login_frame.grid()
		self.root.title("Card Game Login")

	def create_account(self):
		account_id = self.register_account_id.get().strip()
		password = self.register_password.get()
		if not self._validate_credentials(account_id, password):
			return

		self.register_button.configure(state="disabled")
		self.register_status.set("Creating account...")
		threading.Thread(
			target=self._create_account_in_background,
			args=(account_id, password),
			daemon=True,
		).start()

	def _create_account_in_background(self, account_id: str, password: str):
		try:
			self.client.connect()
			result = self.client.create_account(account_id, password)
			self.root.after(0, self._show_create_account_result, result)
		except Exception as error:
			self.root.after(0, self._show_create_account_error, error)

	def _show_create_account_result(self, result: object):
		self.register_button.configure(state="normal")
		if isinstance(result, dict) and result.get("resultCode") == 0:
			self.client.close()
			self.register_account_id.set("")
			self.register_password.set("")
			self.register_status.set("")
			self.account_id.set(result.get("accountId", ""))
			self.show_login_frame()
			messagebox.showinfo("Create account", "Account created successfully.")
			return

		self.client.close()
		self.register_password.set("")
		self.register_status.set("Create account failed")
		messagebox.showerror(
			"Create account failed",
			self._get_result_error(result),
		)

	def _show_create_account_error(self, error: Exception):
		self.client.close()
		self.register_button.configure(state="normal")
		self.register_status.set("Create account failed")
		messagebox.showerror("Create account error", str(error))

	def _get_result_error(self, result: object) -> str:
		if not isinstance(result, dict):
			return str(result)
		return f"Server rejected the request. resultCode={result.get('resultCode')}"

	def login(self):
		if self.login_in_progress:
			return

		account_id = self.account_id.get().strip()
		password = self.password.get()
		if not self._validate_credentials(account_id, password):
			return

		self.login_in_progress = True
		self.login_button.configure(state="disabled")
		self.status.set("Connecting...")
		threading.Thread(
			target=self._login_in_background,
			args=(account_id, password),
			daemon=True,
		).start()

	def _validate_credentials(self, account_id: str, password: str) -> bool:
		if not account_id or not password:
			messagebox.showwarning(
				"Invalid input",
				"Please enter both account ID and password.",
			)
			return False
		if not ACCOUNT_ID_PATTERN.fullmatch(account_id):
			messagebox.showwarning(
				"Invalid account ID",
				"Account ID can contain English letters and numbers only.",
			)
			return False
		if len(password) < 8 or len(password) > 16:
			messagebox.showwarning(
				"Invalid password",
				"Password must be between 8 and 16 characters.",
			)
			return False
		if not PASSWORD_PATTERN.fullmatch(password):
			messagebox.showwarning(
				"Invalid password",
				"Password contains unsupported characters.",
			)
			return False
		return True

	def _login_in_background(self, account_id: str, password: str):
		try:
			self.client.connect()
			result = self.client.login(account_id, password)
			self.root.after(0, self._show_login_result, result)
		except Exception as error:
			self.root.after(0, self._show_login_error, error)

	def _show_login_result(self, result: object):
		self.login_in_progress = False
		self.login_button.configure(state="normal")
		if isinstance(result, dict) and result.get("resultCode") == 0:
			self.status.set("Login successful")
			self._show_game_frame()
			self.uid = result.get("uid")
			self._append_result("Login", result)
			return

		self.status.set("Login failed")
		self.client.close()
		messagebox.showerror("Login failed", json.dumps(result, ensure_ascii=False))

	def _show_login_error(self, error: Exception):
		self.login_in_progress = False
		self.login_button.configure(state="normal")
		self.status.set("Connection failed")
		self.client.close()
		messagebox.showerror("Login error", str(error))

	def _show_game_frame(self):
		self.login_frame.grid_remove()
		self.root.title("Card Game - Main")

		self.game_frame = ttk.Frame(self.root, padding=16)
		self.game_frame.grid(row=0, column=0)

		ttk.Label(
			self.game_frame,
			text=f"Logged in as: {self.account_id.get().strip()}",
		).grid(row=0, column=0, columnspan=2, pady=(0, 12))

		self.summon_button = ttk.Button(
			self.game_frame,
			text="Summon",
			command=self.summon,
		)
		self.summon_button.grid(row=1, column=0, padx=(0, 8), pady=6)

		ttk.Button(
			self.game_frame,
			text="Logout",
			command=self.logout,
		).grid(row=1, column=1, padx=(8, 0), pady=6)

		ttk.Label(self.game_frame, textvariable=self.game_status).grid(
			row=2,
			column=0,
			columnspan=2,
			pady=6,
		)

		self.result_text = tk.Text(
			self.game_frame,
			width=52,
			height=12,
			state="disabled",
			wrap="word",
		)
		self.result_text.grid(row=3, column=0, columnspan=2, pady=(8, 0))

	def summon(self):
		if self.action_in_progress:
			return

		self.action_in_progress = True
		self.summon_button.configure(state="disabled")
		self.game_status.set("Summoning...")
		threading.Thread(
			target=self._summon_in_background,
			daemon=True,
		).start()

	def _summon_in_background(self):
		try:
			result = self.client.summon(self.uid,1, "Normal")
			self.root.after(0, self._handle_summon_result, result)
		except Exception as error:
			self.root.after(0, self._show_summon_error, error)

	def _handle_summon_result(self, result: object):
		self.action_in_progress = False
		self.summon_button.configure(state="normal")
		self.game_status.set("Ready")
		self._append_result("Summon", result)

	def _show_summon_error(self, error: Exception):
		self.action_in_progress = False
		self.summon_button.configure(state="normal")
		self.game_status.set("Summon failed")
		self._append_result("Summon error", str(error))

	def _handle_server_push(self, data: object):
		self.root.after(0, self._append_result, "Server push: Summon.Notify", data)

	def _append_result(self, title: str, result: object):
		self.result_text.configure(state="normal")
		self.result_text.insert(
			"end",
			f"{title}: {json.dumps(result, ensure_ascii=False)}\n",
		)
		self.result_text.see("end")
		self.result_text.configure(state="disabled")

	def logout(self):
		if self.action_in_progress:
			return

		self.client.close()
		self.account_id.set("")
		self.password.set("")
		self.status.set("Disconnected")
		self.game_status.set("Ready")
		self.game_frame.destroy()
		self.game_frame = None
		self.login_frame.grid()
		self.root.title("Card Game Login")

	def close(self):
		self.client.close()
		self.root.destroy()


if __name__ == "__main__":
	root = tk.Tk()
	GameClientWindow(root)
	root.mainloop()