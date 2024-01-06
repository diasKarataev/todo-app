import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './App.css';

const App = () => {
  const [tasks, setTasks] = useState([]);
  const [newTask, setNewTask] = useState({ name: '', details: '' });
  const [editingTask, setEditingTask] = useState(null);

  useEffect(() => {
    const fetchTasks = async () => {
      try {
        const response = await axios.get('http://localhost:8000/tasks');
        setTasks(response.data);
      } catch (error) {
        console.error('Ошибка получения задач:', error);
      }
    };

    fetchTasks();
  }, []);

  const handleInputChange = (event) => {
    const { name, value } = event.target;
    setNewTask({ ...newTask, [name]: value });
  };
  const handleSubmit = async (event) => {
    event.preventDefault();

    try {
      const response = await axios.post('http://localhost:8000/tasks', newTask);
      setTasks([...tasks, response.data]);
      setNewTask({ name: '', details: '' });
    } catch (error) {
      console.error('Ошибка создания задачи:', error);
    }
  };
  const handleEdit = (task) => {
    setEditingTask(task); // Устанавливаем задачу для редактирования
  };

  const handleUpdate = async () => {
    try {
      const response = await axios.put(`http://localhost:8000/tasks/${editingTask.ID}`, editingTask);
      const updatedTasks = tasks.map((task) => (task.ID === editingTask.ID ? response.data : task));
      setTasks(updatedTasks);
      setEditingTask(null); // Сбрасываем редактирование
    } catch (error) {
      console.error('Ошибка обновления задачи:', error);
    }
  };

  const handleEditedInputChange = (event) => {
    const { name, value } = event.target;
    setEditingTask({ ...editingTask, [name]: value });
  };

  const handleCancelEdit = () => {
    setEditingTask(null); // Отменяем редактирование
  };




  const handleDelete = async (taskId) => {
    try {
      await axios.delete(`http://localhost:8000/tasks/${taskId}`);
      const updatedTasks = tasks.filter((task) => task.ID !== taskId);
      setTasks(updatedTasks);
    } catch (error) {
      console.error('Ошибка удаления задачи:', error);
    }
  };

  return (
      <div className='container'>
        <h1>Список задач</h1>
        <form onSubmit={handleSubmit}>
          <input
              type="text"
              name="name"
              placeholder="Название задачи"
              value={newTask.name}
              onChange={handleInputChange}
          />
          <input
              type="text"
              name="details"
              placeholder="Детали"
              value={newTask.details}
              onChange={handleInputChange}
          />
          <button type="submit">Добавить задачу</button>
        </form>
        <ul>
          {tasks.map((task) => (
              <li key={task.ID}>
                {editingTask && editingTask.ID === task.ID ? (
                    <div>
                      <input type="text" name="name" value={editingTask.name} onChange={handleEditedInputChange}/>
                      <input type="text" name="details" value={editingTask.details} onChange={handleEditedInputChange}/>
                      <button onClick={handleUpdate}>Сохранить</button>
                      <button className='cancel-btn' onClick={handleCancelEdit}>Отмена</button>
                    </div>
                ) : (
                    <div>
                      <h2>{task.name}</h2>
                      <p>{task.details}</p>
                      <button onClick={() => handleEdit(task)}>Редактировать</button>
                      <button className='delete-btn' onClick={() => handleDelete(task.ID)}>Удалить</button>
                    </div>
                )}
              </li>
          ))}
        </ul>
      </div>
  );
};

export default App;
