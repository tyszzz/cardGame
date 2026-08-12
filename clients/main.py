import json
import threading
import tkinter as tk
from tkinter import messagebox, ttk

from pitaya_client import PitayaClient

HOST = "172.28.189.203"
PORT = 8080

class LoginWindow:
	def __init__(self, root: tk.Tk):
		self.root = root
		self.root.title("Card Game Login")
		self.root.resizable(False, False)

		self.client = PitayaClient(f"ws://{HOST}:{PORT}/")
		self.login_in_progress = False

		self.account_id = tk.StringVar()
		self.password = tk.StringVar()
		self.status = tk.StringVar(value="Disconnected")

		self._build_ui()
		self.root.protocol("WM_DELETE_WINDOW", self.close)

	def _build_ui(self):
		frame = ttk.Frame(self.root, padding=16)
		frame.grid(row=0, column=0)

		ttk.Label(frame, text="Account").grid(
			row=0, column=0, padx=(0, 8), pady=6, sticky="w"
		)
		account_entry = ttk.Entry(frame, textvariable=self.account_id, width=28)
		account_entry.grid(row=0, column=1, pady=6)

		ttk.Label(frame, text="Password").grid(
			row=1, column=0, padx=(0, 8), pady=6, sticky="w"
		)
		password_entry = ttk.Entry(
			frame,
			textvariable=self.password,
			width=28,
			show="*",
		)
		password_entry.grid(row=1, column=1, pady=6)

		self.login_button = ttk.Button(
			frame,
			text="Login",
			command=self.login,
		)
		self.login_button.grid(row=2, column=0, columnspan=2, pady=(12, 6))

		ttk.Label(frame, textvariable=self.status).grid(
			row=3, column=0, columnspan=2, pady=6
		)

		account_entry.focus_set()
		password_entry.bind("<Return>", lambda _event: self.login())

	def login(self):
		if self.login_in_progress:
			return

		account_id = self.account_id.get().strip()
		password = self.password.get()
		if not account_id or not password:
			messagebox.showwarning(
				"Missing input",
				"Please enter both account and password.",
			)
			return

		self.login_in_progress = True
		self.login_button.configure(state="disabled")
		self.status.set("Connecting...")

		threading.Thread(
			target=self._login_in_background,
			args=(account_id, password),
			daemon=True,
		).start()

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
		else:
			self.status.set("Login failed")

		messagebox.showinfo("Login result", json.dumps(result, ensure_ascii=False))

	def _show_login_error(self, error: Exception):
		self.login_in_progress = False
		self.login_button.configure(state="normal")
		self.status.set("Connection failed")
		messagebox.showerror("Login error", str(error))

	def close(self):
		self.client.close()
		self.root.destroy()


if __name__ == "__main__":
	root = tk.Tk()
	LoginWindow(root)
	root.mainloop()